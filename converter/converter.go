package converter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/textproto"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ConvertToUTF8 reads email content from r, detects its charset, and converts it to UTF-8.
// It returns a new io.Reader containing the UTF-8 encoded content.
func ConvertToUTF8(r io.Reader) (io.Reader, error) {
	contentBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read email content for conversion: %w", err)
	}

	// Create a textproto.Reader to parse headers reliably
	// We need to read the headers first to detect charset, then the body.
	// Since io.Reader can only be read once, we use a bytes.Reader for the header parsing
	// and then reset it or create a new one for the body.
	// For simplicity and to avoid re-reading the entire content, we'll parse headers
	// from a copy of contentBytes, and then use the original contentBytes for body decoding.
	tr := textproto.NewReader(bufio.NewReader(bytes.NewReader(contentBytes)))
	mimeHeader, err := tr.ReadMIMEHeader()
	if err != nil && err != io.EOF { // EOF is expected if there's no body after headers
		// log.Printf("Warning: Converter - Failed to read MIME header: %v", err)
		// Proceed without charset detection if header reading fails
	}

	var detectedCharset string
	contentType := mimeHeader.Get("Content-Type")
	if contentType != "" {
		_, params, err := mime.ParseMediaType(contentType)
		if err == nil {
			if charset, ok := params["charset"]; ok {
				detectedCharset = strings.ToLower(charset)
				// log.Printf("DEBUG: Converter - Detected charset from Content-Type: '%s'", detectedCharset)
			}
		} else {
			// log.Printf("Warning: Converter - Failed to parse Content-Type media type: %v", err)
		}
	}

	// If no charset detected or it's already UTF-8, return original content
	if detectedCharset == "" || detectedCharset == "utf-8" {
		// log.Printf("DEBUG: Converter - No charset detected or already UTF-8. Returning original content.")
		return bytes.NewReader(contentBytes), nil
	}

	// Perform conversion if a supported non-UTF-8 charset is detected
	var decoder encoding.Encoding
	switch detectedCharset {
	case "iso-2022-jp":
		decoder = japanese.ISO2022JP
	case "shift_jis", "shift-jis":
		decoder = japanese.ShiftJIS
	case "euc-jp", "euc_jp":
		decoder = japanese.EUCJP
	default:
		// log.Printf("Warning: Converter - Detected unsupported charset '%s' for conversion. Proceeding without conversion.", detectedCharset)
		return bytes.NewReader(contentBytes), nil // Return original if unsupported
	}

	// log.Printf("DEBUG: Converter - Converting from '%s' to UTF-8.", detectedCharset)
	decodedBodyBytes, _, err := transform.Bytes(decoder.NewDecoder(), contentBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode content from %s: %w", detectedCharset, err)
	}

	// Reconstruct headers with updated Content-Type charset
	if detectedCharset != "" && detectedCharset != "utf-8" {
		// Update charset in Content-Type header
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err == nil {
			params["charset"] = "UTF-8"
			mimeHeader.Set("Content-Type", mime.FormatMediaType(mediaType, params))
			// log.Printf("DEBUG: Converter - Rewrote Content-Type charset to UTF-8 in header.")
		} else {
			// log.Printf("Warning: Converter - Failed to re-parse Content-Type for rewriting: %v", err)
		}
	}

	// Reconstruct the full email with updated headers and decoded body
	var buf bytes.Buffer
	// Write headers
	for key, values := range mimeHeader {
		for _, value := range values {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	buf.WriteString("\r\n") // End of headers
	buf.Write(decodedBodyBytes)

	return bytes.NewReader(buf.Bytes()), nil
}