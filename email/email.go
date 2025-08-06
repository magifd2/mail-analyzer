package email

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ParsedEmail holds the extracted information from an email.
type ParsedEmail struct {
	MessageID string
	From      []*mail.Address
	To        []*mail.Address
	Subject   string
	Body      string
	URLs      []string
	Header    mail.Header
}

// Parse reads an email from an io.Reader and extracts key information.
func Parse(r io.Reader) (*ParsedEmail, error) {
	entity, err := message.Read(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read message entity: %w", err)
	}

	mr := mail.NewReader(entity)

	header := mr.Header
	from, _ := header.AddressList("From")
	to, _ := header.AddressList("To")
	subject, _ := header.Subject()
	messageID, _ := header.MessageID()

	body, urls, err := extractBodyAndURLs(entity)
	if err != nil {
		return nil, err
	}

	return &ParsedEmail{
		MessageID: strings.Trim(messageID, "<> "),
		From:      from,
		To:        to,
		Subject:   subject,
		Body:      body,
		URLs:      urls,
		Header:    header,
	}, nil
}

func extractBodyAndURLs(entity *message.Entity) (string, []string, error) {
	mediaType, params, err := entity.Header.ContentType()
	if err != nil {
		mediaType = "text/plain"
		params = make(map[string]string)
	}

	var bodyBuilder strings.Builder
	var urls []string

	hrefRegex := regexp.MustCompile(`href\s*=\s*["'](https?://[^"]+)["']`)
	urlRegex := regexp.MustCompile(`https?://[^\s"<>]*[^\s"<>,.?!;)]`)

	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			content, _ := io.ReadAll(entity.Body)
			bodyBuilder.WriteString(string(content))
		} else {
			mr := multipart.NewReader(entity.Body, boundary)
			for {
				part, err := mr.NextPart()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					log.Printf("Warning: could not read multipart part: %v", err)
					continue
				}
				defer part.Close()

				partMediaType, partParams, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
				if err != nil {
					log.Printf("Warning: could not parse content type of multipart part: %v", err)
					continue
				}

				partContent, err := io.ReadAll(part)
				if err != nil {
					log.Printf("Warning: could not read content of multipart part: %v", err)
					continue
				}

				// Decode charset if specified
			charset := partParams["charset"]
			if charset != "" {
				log.Printf("DEBUG: Decoding part with charset: %s", charset)
					decodedContent, decodeErr := decodeCharset(partContent, charset)
					if decodeErr == nil {
						partContent = decodedContent
					} else {
						log.Printf("Warning: Failed to decode charset %s: %v", charset, decodeErr)
					}
				}

				partBodyText := string(partContent)

				if partMediaType == "text/html" || partMediaType == "text/plain" {
					hrefMatches := hrefRegex.FindAllStringSubmatch(partBodyText, -1)
					for _, match := range hrefMatches {
						if len(match) > 1 {
							urls = append(urls, match[1])
						}
					}

					if partMediaType == "text/html" {
						re := regexp.MustCompile(`<.*?>`)
						partBodyText = re.ReplaceAllString(partBodyText, " ")
					}

					foundUrls := urlRegex.FindAllString(partBodyText, -1)
					urls = append(urls, foundUrls...)

					bodyBuilder.WriteString(partBodyText)
					bodyBuilder.WriteString("\n")
				}
			}
		}
	} else if mediaType == "text/plain" || mediaType == "text/html" {
		content, err := io.ReadAll(entity.Body)
			if err != nil {
				return "", nil, err
			}

			// Decode charset if specified
			charset := params["charset"]
			if charset != "" {
				log.Printf("DEBUG: Decoding main body with charset: %s", charset)
				decodedContent, decodeErr := decodeCharset(content, charset)
				if decodeErr == nil {
					content = decodedContent
				} else {
					log.Printf("Warning: Failed to decode charset %s: %v", charset, decodeErr)
				}
			}

			bodyText := string(content)

			hrefMatches := hrefRegex.FindAllStringSubmatch(bodyText, -1)
			for _, match := range hrefMatches {
				if len(match) > 1 {
					urls = append(urls, match[1])
				}
			}

			if mediaType == "text/html" {
				re := regexp.MustCompile(`<.*?>`)
				bodyText = re.ReplaceAllString(bodyText, " ")
			}

			foundUrls := urlRegex.FindAllString(bodyText, -1)
			urls = append(urls, foundUrls...)

			bodyBuilder.WriteString(bodyText)
	}

	uniqueUrls := make(map[string]bool)
	var resultUrls []string
	for _, u := range urls {
		u = strings.TrimRight(u, ".?!,;)")
		if !uniqueUrls[u] {
			uniqueUrls[u] = true
			resultUrls = append(resultUrls, u)
		}
	}

	return strings.TrimSpace(bodyBuilder.String()), resultUrls, nil
}

// decodeCharset decodes content from a given charset to UTF-8.
func decodeCharset(content []byte, charset string) ([]byte, error) {
	charset = strings.ToLower(charset)

	var decoder *encoding.Decoder
	switch charset {
	case "iso-2022-jp":
		decoder = japanese.ISO2022JP.NewDecoder()
	case "shift_jis", "shift-jis":
		decoder = japanese.ShiftJIS.NewDecoder()
	case "euc-jp", "euc_jp":
		decoder = japanese.EUCJP.NewDecoder()
	case "utf-8":
		return content, nil // Already UTF-8
	default:
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}

	decoded, _, err := transform.Bytes(decoder, content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode content from %s: %w", charset, err)
	}
	return decoded, nil
}
