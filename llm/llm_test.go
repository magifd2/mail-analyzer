package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"mail-analyzer/config"
)

func TestOpenAIProvider_AnalyzeText(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   APIResponse
		mockStatusCode int
		prompt         string
		tools          []APITool
		want           *Judgment
		wantErr        bool
	}{
		{
			name: "Successful analysis with tool call",
			mockResponse: APIResponse{
				Choices: []Choice{
					{
						Message: Message{
							ToolCalls: []ToolCall{
								{
									Function: FunctionCall{
										Arguments: `{"is_suspicious": true, "category": "Phishing", "reason": "Contains a suspicious link.", "confidence_score": 0.9}`,
									},
								},
							},
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			prompt:         "Analyze this email.",
			want: &Judgment{
				IsSuspicious:    true,
				Category:        "Phishing",
				Reason:          "Contains a suspicious link.",
				ConfidenceScore: 0.9,
			},
			wantErr: false,
		},
		{
			name:           "API returns an error",
			mockResponse:   APIResponse{Error: &APIError{Message: "Internal server error"}},
			mockStatusCode: http.StatusInternalServerError,
			prompt:         "Analyze this email.",
			want:           nil,
			wantErr:        true,
		},
		{
			name:           "No tool calls in response",
			mockResponse:   APIResponse{Choices: []Choice{{Message: Message{}}}},
			mockStatusCode: http.StatusOK,
			prompt:         "Analyze this email.",
			want:           nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			cfg := &config.Config{
				OpenAIAPIKey:  "test-key",
				OpenAIBaseURL: server.URL,
				ModelName:     "test-model",
			}

			provider := NewOpenAIProvider(cfg)
			got, err := provider.AnalyzeText(context.Background(), tt.prompt, tt.tools, "")

			if (err != nil) != tt.wantErr {
				t.Errorf("OpenAIProvider.AnalyzeText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OpenAIProvider.AnalyzeText() = %v, want %v", got, tt.want)
			}
		})
	}
}
