package analyzer

import (
	"context"
	"fmt"
	"strings"

	"mail-analyzer/email"
	"mail-analyzer/llm"
)

// LLMProvider defines the interface for a Large Language Model provider.
type LLMProvider interface {
	AnalyzeText(ctx context.Context, prompt string, tools []llm.APITool, toolChoice string) (*llm.Judgment, error)
}

// EmailAnalyzer is responsible for analyzing emails.
type EmailAnalyzer struct {
	provider LLMProvider
}

// NewEmailAnalyzer creates a new EmailAnalyzer.
func NewEmailAnalyzer(provider LLMProvider) *EmailAnalyzer {
	return &EmailAnalyzer{provider: provider}
}

// Analyze performs the analysis of a single email.
func (a *EmailAnalyzer) Analyze(ctx context.Context, email *email.ParsedEmail) (*llm.Judgment, error) {
	prompt := buildPrompt(email)
	tool := getAnalysisTool()
	return a.provider.AnalyzeText(ctx, prompt, []llm.APITool{tool}, "auto")
}

func buildPrompt(email *email.ParsedEmail) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("Please analyze the following email and determine if it is safe, spam, or phishing.\n\n")
	promptBuilder.WriteString("--- Email Headers ---\n")
	if len(email.From) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("From: %s\n", email.From[0].String()))
	}
	if len(email.To) > 0 {
		var toAddresses []string
		for _, addr := range email.To {
			toAddresses = append(toAddresses, addr.String())
		}
		promptBuilder.WriteString(fmt.Sprintf("To: %s\n", strings.Join(toAddresses, ", ")))
	}
	promptBuilder.WriteString(fmt.Sprintf("Subject: %s\n", email.Subject))
	if returnPath, err := email.Header.Text("Return-Path"); err == nil {
		promptBuilder.WriteString(fmt.Sprintf("Return-Path: %s\n", returnPath))
	}
	if replyTo, err := email.Header.AddressList("Reply-To"); err == nil {
		var replyToAddresses []string
		for _, addr := range replyTo {
			replyToAddresses = append(replyToAddresses, addr.String())
		}
		promptBuilder.WriteString(fmt.Sprintf("Reply-To: %s\n", strings.Join(replyToAddresses, ", ")))
	}

	promptBuilder.WriteString("\n--- Email Body ---\n")
	body := email.Body
	if len(body) > 4000 { // Truncate long bodies
		body = body[:4000] + "\n... (truncated)"
	}
	promptBuilder.WriteString(body)

	promptBuilder.WriteString("\n\n--- Extracted URLs---\n")
	if len(email.URLs) > 0 {
		for _, u := range email.URLs {
			promptBuilder.WriteString(u + "\n")
		}
	} else {
		promptBuilder.WriteString("No URLs found.\n")
	}

	promptBuilder.WriteString("\n--- Analysis Instructions---\n")
	promptBuilder.WriteString("Based on all the information above, call the 'report_analysis_result' function with your conclusion.")

	return promptBuilder.String()
}

func getAnalysisTool() llm.APITool {
	return llm.APITool{
		Type: "function",
		Function: llm.APIFunctionDef{
			Name:        "report_analysis_result",
			Description: "Reports the analysis result of an email.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"is_suspicious":     map[string]any{"type": "boolean", "description": "Whether the email is suspicious (phishing, spam, etc.)."},
					"category":          map[string]any{"type": "string", "enum": []string{"Phishing", "Spam", "Safe"}, "description": "The category of the email."},
					"reason":            map[string]any{"type": "string", "description": "A brief explanation for the judgment."}, 
					"confidence_score":  map[string]any{"type": "number", "description": "Confidence score of the analysis from 0.0 to 1.0."},
				},
				"required": []string{"is_suspicious", "category", "reason", "confidence_score"},
			},
		},
	}
}