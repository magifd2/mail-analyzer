package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"mail-analyzer/config"
)

// --- Struct Definitions ---

// Judgment is the structured analysis result from the LLM.
type Judgment struct {
	IsSuspicious    bool    `json:"is_suspicious"`
	Category        string  `json:"category"` // e.g., "Phishing", "Spam", "Safe"
	Reason          string  `json:"reason"`
	ConfidenceScore float64 `json:"confidence_score"`
}

// --- LLM API Related Structs ---

type APIRequest struct {
	Model      string    `json:"model"`
	Messages   []Message `json:"messages"`
	Tools      []APITool `json:"tools,omitempty"`
	ToolChoice any       `json:"tool_choice,omitempty"`
}

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type APITool struct {
	Type     string         `json:"type"`
	Function APIFunctionDef `json:"function"`
}

type APIFunctionDef struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters"`
}

type APIResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ToolCall struct {
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Arguments string `json:"arguments"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// --- Provider Implementation ---

// OpenAIProvider implements the analyzer.LLMProvider interface using the OpenAI API.
type OpenAIProvider struct {
	client  *http.Client
	config  *config.Config
	baseURL string
}

// NewOpenAIProvider creates a new OpenAIProvider.
func NewOpenAIProvider(cfg *config.Config) *OpenAIProvider {
	return &OpenAIProvider{
		client: &http.Client{
			Timeout: 90 * time.Second,
		},
		config:  cfg,
		baseURL: cfg.OpenAIBaseURL,
	}
}

// AnalyzeText sends the prompt to the OpenAI API and returns the structured judgment.
func (p *OpenAIProvider) AnalyzeText(ctx context.Context, prompt string, tools []APITool, toolChoice string) (*Judgment, error) {
	messages := []Message{
		{Role: "system", Content: "You are a senior cybersecurity analyst specializing in email threat detection. Analyze the provided email data and use the specified tool to report your findings."},
		{Role: "user", Content: prompt},
	}

	apiRequest := APIRequest{
		Model:    p.config.ModelName,
		Messages: messages,
		Tools:    tools,
	}

	if toolChoice != "" {
		apiRequest.ToolChoice = toolChoice
	}

	reqBody, err := json.Marshal(apiRequest)
	if err != nil {
		return nil, fmt.Errorf("could not marshal API request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("could not create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.OpenAIAPIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResponse APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("could not decode API response: %w", err)
	}

	if apiResponse.Error != nil {
		return nil, fmt.Errorf("API error: [%s] %s", apiResponse.Error.Code, apiResponse.Error.Message)
	}

	if len(apiResponse.Choices) == 0 || len(apiResponse.Choices[0].Message.ToolCalls) == 0 {
		return nil, errors.New("API did not return a valid tool call")
	}

	toolCallArgs := apiResponse.Choices[0].Message.ToolCalls[0].Function.Arguments
	var judgment Judgment
	if err := json.Unmarshal([]byte(toolCallArgs), &judgment); err != nil {
		return nil, fmt.Errorf("could not unmarshal tool call arguments: %w", err)
	}

	return &judgment, nil
}
