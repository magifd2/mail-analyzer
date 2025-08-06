package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/emersion/go-mbox"
	"github.com/emersion/go-message/mail"
	"mail-analyzer/analyzer"
	"mail-analyzer/config"
	"mail-analyzer/email"
	"mail-analyzer/llm"
)

// FinalOutput is the final JSON output structure.
type FinalOutput struct {
	SourceFile      string            `json:"source_file"`
	AnalysisResults []*AnalysisResult `json:"analysis_results"`
}

// AnalysisResult is the result for a single email.
type AnalysisResult struct {
	MessageID string         `json:"message_id"`
	Subject   string         `json:"subject"`
	From      []string       `json:"from"`
	To        []string       `json:"to"`
	Judgment  *llm.Judgment  `json:"judgment"`
}

func main() {
	// 1. Load configuration
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <path/to/your/mail.mbox> [config.json]")
	}
	mboxPath := os.Args[1]
	configPath := ""
	if len(os.Args) > 2 {
		configPath = os.Args[2]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	if cfg.OpenAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY must be set in config file or environment variable.")
	}

	// 2. Setup analyzer
	llmProvider := llm.NewOpenAIProvider(cfg)
	emailAnalyzer := analyzer.NewEmailAnalyzer(llmProvider)

	// 3. Open and read mbox file
	file, err := os.Open(mboxPath)
	if err != nil {
		log.Fatalf("Error opening mbox file: %v", err)
	}
	defer file.Close()

	// 4. Process each message
	var results []*AnalysisResult
	mboxReader := mbox.NewReader(file)
	for {
		rawMessage, err := mboxReader.NextMessage()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Warning: Could not read message: %v", err)
			continue
		}

		parsedEmail, err := email.Parse(rawMessage)
		if err != nil {
			log.Printf("Warning: Failed to parse email: %v", err)
			continue
		}

		judgment, err := emailAnalyzer.Analyze(context.Background(), parsedEmail)
		if err != nil {
			log.Printf("Warning: Failed to analyze email (Message-ID: %s): %v", parsedEmail.MessageID, err)
			continue
		}

		results = append(results, &AnalysisResult{
			MessageID: parsedEmail.MessageID,
			Subject:   parsedEmail.Subject,
			From:      convertAddresses(parsedEmail.From),
			To:        convertAddresses(parsedEmail.To),
			Judgment:  judgment,
		})
	}

	// 5. Output results as JSON
	output := FinalOutput{
		SourceFile:      mboxPath,
		AnalysisResults: results,
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	fmt.Println(string(jsonOutput))
}

func convertAddresses(addresses []*mail.Address) []string {
	var result []string
	for _, addr := range addresses {
		result = append(result, addr.String())
	}
	return result
}