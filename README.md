# AI-Powered Suspicious Email Analyzer

A command-line tool that analyzes a single email from an `.eml` file, using a Large Language Model (LLM) to detect suspicious content like phishing or spam. It outputs a structured JSON report for the email.

This tool is designed for flexibility, allowing it to be easily integrated into automated workflows and security analysis pipelines.

## Features

-   **EML File Analysis**: Analyzes a single email directly from a standard `.eml` file.
-   **Automatic Charset Handling**: Automatically detects and converts various email charsets (including `iso-2022-jp`) to UTF-8 for seamless processing.
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

The tool automatically looks for `config.json` in `~/.config/mail-analyzer/`. If you provide a path via the command line, that path will be used instead.

**Directory:**
```sh
mkdir -p ~/.config/mail-analyzer
```

**File (`~/.config/mail-analyzer/config.json`):**
```json
{
  "openai_api_key": "sk-your_openai_api_key_here",
  "openai_api_base_url": "https://api.openai.com/v1/chat/completions",
  "model_name": "gpt-4-turbo"
}
```
-   `openai_api_key` (Required): Your API key for the LLM service.
-   `openai_api_base_url` (Optional): The base URL of the OpenAI-compatible API.
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

Provide the path to your `.eml` file and optionally a path to your `config.json`.

```go
# Using go run
go run main.go /path/to/your/email.eml /path/to/your/config.json

# Using the compiled binary
./mail-analyzer /path/to/your/email.eml /path/to/your/config.json
```

### Analyze from Standard Input

You can also pipe the content of an `.eml` file directly to `mail-analyzer`. This is useful for integrating with other tools or scripts.

```sh
cat /path/to/your/email.eml | ./mail-analyzer
```

### Debugging

To enable debug logging (output to stderr), use the `--debug` or `-d` flag:

```sh
./mail-analyzer --debug /path/to/your/email.eml
cat /path/to/your/email.eml | ./mail-analyzer -d
```

---

## Output Format

The tool outputs a single JSON object to standard output containing the analysis results for all emails.

**Example Output:**
```json
{
  "source_file": "/path/to/your/email.eml",
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

```