# AI-Powered Suspicious Email Analyzer

A command-line tool that analyzes emails from an `mbox` file, using a Large Language Model (LLM) to detect suspicious content like phishing or spam. It outputs a structured JSON report for each email.

This tool is designed for flexibility, allowing it to be easily integrated into automated workflows and security analysis pipelines.

## Features

-   **Mbox File Analysis**: Analyzes emails directly from a standard `mbox` file.
-   **LLM-Powered Analysis**: Leverages any OpenAI-compatible API with Tool-Calling capabilities for intelligent and structured email analysis.
-   **Structured JSON Output**: Provides analysis results in a clean, machine-readable format.
-   **Flexible Configuration**: Configure via a JSON file and/or environment variables.
-   **Built with Go**: A single, fast, and portable binary.

---

## Prerequisites

-   [Go](https://go.dev/doc/install) (version 1.24 or later)
-   An API key for an OpenAI-compatible service.

---

## Installation

1.  Clone the repository:
    ```sh
    git clone https://your-repository-url/mail-analyzer.git
    cd mail-analyzer
    ```

2.  Install dependencies:
    ```sh
    go mod tidy
    ```

3.  Build the binary using the Makefile:
    ```sh
    make build
    ```
    This will create a `mail-analyzer` executable in the project root.

---

## Configuration

The tool loads its configuration with the following priority: **Environment Variables > Configuration File > Defaults**.

### 1. Configuration File (Recommended)

Create a `config.json` file in the project root or any other location.

**Example `config.json`:**
```json
{
  "openai_api_key": "sk-your_openai_api_key_here",
  "openai_base_url": "https://api.openai.com/v1/chat/completions",
  "model_name": "gpt-4-turbo"
}
```
-   `openai_api_key` (Required): Your API key for the LLM service.
-   `openai_base_url` (Optional): The base URL of the OpenAI-compatible API.
-   `model_name` (Optional): The model to use for analysis. Defaults to `gpt-4-turbo`.

### 2. Environment Variables

You can override any setting from the configuration file by setting environment variables. This is useful for CI/CD environments or temporary changes.

```sh
export OPENAI_API_KEY="sk-your_api_key"
export OPENAI_API_BASE_URL="https://your-custom-proxy.com/v1/chat/completions"
export MODEL_NAME="your-custom-model"
```

---

## Usage

You can run the tool directly using `go run` or the compiled binary.

### Analyze from a File

Provide the path to your `mbox` file and optionally a path to your `config.json`.

```sh
# Using go run

```go
go run main.go /path/to/your/emails.mbox /path/to/your/config.json
```

# Using the compiled binary

```go
./mail-analyzer /path/to/your/emails.mbox /path/to/your/config.json
```

If the config path is omitted, the tool will look for environment variables.

---

## Output Format

The tool outputs a single JSON object to standard output containing the analysis results for all emails.

**Example Output:**
```json
{
  "source_file": "/path/to/your/emails.mbox",
  "analysis_results": [
    {
      "message_id": "<phishing-example-id@mail.example.com>",
      "subject": "Urgent: Verify Your Account Now!",
      "from": [
        "\"Suspicious Bank\" <security@suspicious-bank.com>"
      ],
      "to": [
        "\"Your Name\" <user@example.com>"
      ],
      "judgment": {
        "is_suspicious": true,
        "category": "Phishing",
        "reason": "The email uses urgent language and contains a suspicious link designed to steal credentials. The sender's domain does not match the official bank's domain.",
        "confidence_score": 0.98
      }
    }
  ]
}
```

---

## For Developers

### Makefile Targets

-   `make build`: Compiles the Go source code and creates the `mail-analyzer` binary.
-   `make test`: Runs all tests in the project.
-   `make clean`: Removes the compiled binary.