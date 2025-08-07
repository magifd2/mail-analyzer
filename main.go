package main

import (
	"context"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

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
	// Setup logging
	debug := flag.Bool("debug", false, "Enable debug logging")
	d := flag.Bool("d", false, "Enable debug logging (shorthand)")
	flag.Parse()

	if !(*debug || *d) {
		log.SetOutput(ioutil.Discard) // Discard all log.Printf output
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile) // Add file and line number to debug logs
	}

	// Adjust os.Args after flag parsing
	args := flag.Args()

	var rawMessage []byte
	var sourceFile string
	var err error

	// 1. Load configuration
	var cfg *config.Config // Declare cfg here
	configPath := ""

	if len(args) < 1 {
		// Read from stdin if no file path is provided
		log.Println("No EML file path provided. Reading from stdin...")
		rawMessage, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
		sourceFile = "stdin" // Indicate source is stdin

		// Determine config path for stdin case
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting user home directory: %v", err)
		}
		configPath = fmt.Sprintf("%s/.config/mail-analyzer/config.json", homeDir)

		cfg, err = config.Load(configPath)
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}
	} else {
		emlPath := args[0]
		if len(args) > 1 {
			configPath = args[1] // configPath is set if provided as second argument
		} else {
			// If only EML path is provided, use default config path
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("Error getting user home directory: %v", err)
			}
			configPath = fmt.Sprintf("%s/.config/mail-analyzer/config.json", homeDir)
		}

		cfg, err = config.Load(configPath)
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}

		// 3. Read eml file
		rawMessage, err = os.ReadFile(emlPath)
		if err != nil {
			log.Fatalf("Error reading eml file: %v", err)
		}
		sourceFile = emlPath
	}

	// Ensure at least one of OpenAIAPIKey or OpenAIAPIBaseURL is set
	// If OpenAIAPIBaseURL is set, APIKey can be empty (for local LLMs)
	if cfg.OpenAIAPIKey == "" && cfg.OpenAIBaseURL == "" {
		log.Fatal("OPENAI_API_KEY or OPENAI_API_BASE_URL must be set in config file or environment variable.")
	}

	// 2. Setup analyzer
	llmProvider := llm.NewOpenAIProvider(cfg)
	emailAnalyzer := analyzer.NewEmailAnalyzer(llmProvider)

	// 4. Process the message
	var results []*AnalysisResult
	parsedEmail, err := email.Parse(bytes.NewReader(rawMessage))
	if err != nil {
		log.Fatalf("Error parsing email: %v", err)
	}

	judgment, err := emailAnalyzer.Analyze(context.Background(), parsedEmail)
	if err != nil {
		log.Fatalf("Error analyzing email (Message-ID: %s): %v", parsedEmail.MessageID, err)
	}

	results = append(results, &AnalysisResult{
		MessageID: parsedEmail.MessageID,
		Subject:   parsedEmail.Subject,
		From:      convertAddresses(parsedEmail.From),
		To:        convertAddresses(parsedEmail.To),
		Judgment:  judgment,
	})

	// 5. Output results as JSON
	output := FinalOutput{
		SourceFile:      sourceFile,
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