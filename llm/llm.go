package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
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

type LLMToolCallResponse struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
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

	// Only set Authorization header if API key is provided
	if p.config.OpenAIAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.OpenAIAPIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read API response body: %w", err)
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(respBody, &apiResponse); err != nil {
		return nil, fmt.Errorf("could not decode API response: %w", err)
	}

	if apiResponse.Error != nil {
		return nil, fmt.Errorf("API error: [%s] %s", apiResponse.Error.Code, apiResponse.Error.Message)
	}

	// --- Custom parsing for local LLM tool calls ---
	// Check if the response contains a message with content that includes tool call markers
	if len(apiResponse.Choices) > 0 && apiResponse.Choices[0].Message.Content != "" {
		content := apiResponse.Choices[0].Message.Content
		log.Printf("DEBUG LLM Response Content: %s", content)
		// Regex to find the JSON string between [TOOL_REQUEST] and [END_TOOL_REQUEST]
		re := regexp.MustCompile(`(?s)\[TOOL_REQUEST\](.*)\[END_TOOL_REQUEST\]`)
		matches := re.FindStringSubmatch(content)

		if len(matches) > 1 {
			toolCallArgs := strings.TrimSpace(matches[1])
			log.Printf("DEBUG Extracted Tool Call Args (from TOOL_REQUEST): %s", toolCallArgs)

			var toolCallResponse LLMToolCallResponse
			if err := json.Unmarshal([]byte(toolCallArgs), &toolCallResponse); err != nil {
				log.Printf("ERROR: Could not unmarshal tool call response from TOOL_REQUEST: %v", err)
				return nil, fmt.Errorf("could not unmarshal tool call response: %w", err)
			}

			var judgment Judgment
			if err := json.Unmarshal([]byte(toolCallResponse.Arguments), &judgment); err != nil {
				log.Printf("ERROR: Could not unmarshal judgment from TOOL_REQUEST arguments: %v", err)
				return nil, fmt.Errorf("could not unmarshal judgment from tool call arguments: %w", err)
			}
			log.Printf("DEBUG: Successfully parsed from TOOL_REQUEST.")
			return &judgment, nil
		}

		// If no TOOL_REQUEST tags, try to parse the entire content as a JSON tool call
		trimmedContent := strings.TrimSpace(content)
		log.Printf("DEBUG Attempting to parse entire content as JSON: %s", trimmedContent)
		var toolCallResponse LLMToolCallResponse
		if err := json.Unmarshal([]byte(trimmedContent), &toolCallResponse); err == nil {
			log.Printf("DEBUG: Successfully unmarshaled entire content to LLMToolCallResponse.")
			var judgment Judgment
			if err := json.Unmarshal([]byte(toolCallResponse.Arguments), &judgment); err == nil {
				log.Printf("DEBUG: Successfully unmarshaled judgment from entire content.")
				return &judgment, nil
			} else {
				log.Printf("ERROR: Could not unmarshal judgment from entire content arguments: %v", err)
			}
		} else {
			log.Printf("ERROR: Could not unmarshal entire content to LLMToolCallResponse: %v", err)
		}
	}

	// Fallback to standard tool_calls field if custom parsing fails or is not applicable
	if len(apiResponse.Choices) > 0 && len(apiResponse.Choices[0].Message.ToolCalls) > 0 {
		toolCallArgs := apiResponse.Choices[0].Message.ToolCalls[0].Function.Arguments
		log.Printf("DEBUG Attempting to parse from standard tool_calls field: %s", toolCallArgs)
		var judgment Judgment
		if err := json.Unmarshal([]byte(toolCallArgs), &judgment); err != nil {
			log.Printf("ERROR: Could not unmarshal tool call arguments from standard field: %v", err)
			return nil, fmt.Errorf("could not unmarshal tool call arguments from standard field: %w", err)
		}
		log.Printf("DEBUG: Successfully parsed from standard tool_calls field.")
		return &judgment, nil
	}

	log.Printf("ERROR: API did not return a valid tool call in expected format. Response: %+v", apiResponse)
	return nil, errors.New("API did not return a valid tool call in expected format")
}