# AI-Powered Suspicious Email Analyzer

A command-line tool that analyzes emails from an `mbox` file or standard input, using a Large Language Model (LLM) to detect suspicious content like phishing or spam. It outputs a structured JSON report for each email.

This tool is designed for flexibility, allowing it to be easily integrated into automated workflows and security analysis pipelines.

## Features

-   **Flexible Input**: Analyzes `mbox` data from a file or standard input (stdin).
-   **LLM-Powered Analysis**: Leverages any OpenAI-compatible API with Tool-Calling capabilities for intelligent email analysis.
-   **Structured Output**: Provides analysis results in a clean, machine-readable JSON format.
-   **Flexible Configuration**: Configure via a JSON file, environment variables, or rely on sensible defaults.
-   **Built with Go**: A single, fast, and portable binary.

---

## Prerequisites

-   [Go](https://go.dev/doc/install) (version 1.18 or later)

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

Create a configuration file at `~/.config/mail-analyzer/config.json`. This is the easiest way to set up the tool for regular use.

**Directory:**
```sh
mkdir -p ~/.config/mail-analyzer
```

**File (`~/.config/mail-analyzer/config.json`):**
```json
{
  "api_key": "sk-your_openai_api_key_here",
  "base_url": "[https://api.openai.com/v1/chat/completions](https://api.openai.com/v1/chat/completions)",
  "model_name": "gpt-4-turbo"
}
```
-   `api_key` (Required): Your API key for the LLM service.
-   `base_url` (Optional): The base URL of the OpenAI-compatible API. Defaults to the official OpenAI endpoint.
-   `model_name` (Optional): The model to use for analysis. Defaults to `gpt-4-turbo`.

### 2. Environment Variables

You can override any setting from the configuration file by setting environment variables. This is useful for CI/CD environments or temporary changes.

```sh
export OPENAI_API_KEY="sk-your_api_key"
export OPENAI_API_BASE_URL="[https://your-custom-proxy.com/v1/chat/completions](https://your-custom-proxy.com/v1/chat/completions)"
export OPENAI_MODEL_NAME="your-custom-model"
```

---

## Usage

You can run the tool directly using `go run` or build a binary for easier access (`go build -o mail-analyzer`).

### Analyze from a File

Provide the path to your `mbox` file as a command-line argument.

```sh
go run main.go /path/to/your/emails.mbox
```

### Analyze from Standard Input (stdin)

If no file path is provided, the tool will read data from standard input. This allows you to pipe data from other commands.

```sh
cat /path/to/your/emails.mbox | go run main.go
```

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
      "from": "\"Suspicious Bank\" <security@suspicious-bank.com>",
      "to": [
        "\"Your Name\" <user@example.com>"
      ],
      "judgment": {
        "is_suspicious": true,
        "category": "Phishing",
        "reason": "The email uses urgent language and contains a suspicious link designed to steal credentials. The sender's domain does not match the official bank's domain.",
        "confidence_score": 0.98
      }
    },
    {
      "message_id": "<safe-example-id@mail.example.org>",
      "subject": "Project Update Meeting",
      "from": "\"Your Colleague\" <colleague@example.com>",
      "to": [
        "\"Your Name\" <user@example.com>"
      ],
      "judgment": {
        "is_suspicious": false,
        "category": "Safe",
        "reason": "This is a legitimate internal communication regarding a scheduled meeting.",
        "confidence_score": 0.99
      }
    }
  ]
}
```
