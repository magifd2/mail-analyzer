package analyzer

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/emersion/go-message/mail"
	"mail-analyzer/email"
	"mail-analyzer/llm"
)

// MockLLMProvider is a mock implementation of the LLMProvider interface for testing.
type MockLLMProvider struct {
	AnalyzeTextFunc func(ctx context.Context, prompt string, tools []llm.APITool, toolChoice string) (*llm.Judgment, error)
}

func (m *MockLLMProvider) AnalyzeText(ctx context.Context, prompt string, tools []llm.APITool, toolChoice string) (*llm.Judgment, error) {
	if m.AnalyzeTextFunc != nil {
		return m.AnalyzeTextFunc(ctx, prompt, tools, toolChoice)
	}
	return nil, errors.New("AnalyzeTextFunc is not implemented")
}

func TestEmailAnalyzer_Analyze(t *testing.T) {
	tests := []struct {
		name        string
		provider    LLMProvider
		parsedEmail *email.ParsedEmail
		want        *llm.Judgment
		wantErr     bool
	}{
		{
			name: "Successful analysis",
			provider: &MockLLMProvider{
				AnalyzeTextFunc: func(ctx context.Context, prompt string, tools []llm.APITool, toolChoice string) (*llm.Judgment, error) {
					// Basic check to ensure the prompt contains key elements
					if !strings.Contains(prompt, "Subject: Special Offer!") || !strings.Contains(prompt, "http://promo.example.com") {
						t.Errorf("AnalyzeText received an unexpected prompt: %s", prompt)
					}
					// Check if the correct tool is being passed
					if len(tools) != 1 || tools[0].Function.Name != "report_analysis_result" {
						t.Errorf("AnalyzeText received incorrect tools: %+v", tools)
					}
					return &llm.Judgment{
						IsSuspicious:    true,
						Category:        "Marketing",
						Reason:          "Promotional content.",
						ConfidenceScore: 0.8,
					}, nil
				},
			},
			parsedEmail: &email.ParsedEmail{
				Subject: "Special Offer!",
				Body:    "Buy now and get 50% off. Visit http://promo.example.com",
				URLs:    []string{"http://promo.example.com"},
				Header:  mail.Header{},
			},
			want: &llm.Judgment{
				IsSuspicious:    true,
				Category:        "Marketing",
				Reason:          "Promotional content.",
				ConfidenceScore: 0.8,
			},
			wantErr: false,
		},
		{
			name: "LLM provider returns an error",
			provider: &MockLLMProvider{
				AnalyzeTextFunc: func(ctx context.Context, prompt string, tools []llm.APITool, toolChoice string) (*llm.Judgment, error) {
					return nil, errors.New("LLM API error")
				},
			},
			parsedEmail: &email.ParsedEmail{Subject: "Error test", Header: mail.Header{}},
			want:        nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewEmailAnalyzer(tt.provider)
			got, err := analyzer.Analyze(context.Background(), tt.parsedEmail)

			if (err != nil) != tt.wantErr {
				t.Errorf("EmailAnalyzer.Analyze() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EmailAnalyzer.Analyze() = %v, want %v", got, tt.want)
			}
		})
	}
}