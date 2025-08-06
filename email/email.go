package email

import (
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
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
		return nil, err
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
	urlRegex := regexp.MustCompile(`https?://[^\s"'<>]*[^\s"'<>,.?!;)]`)

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

				partMediaType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
				if err != nil {
					log.Printf("Warning: could not parse content type of multipart part: %v", err)
					continue
				}

				partContent, err := io.ReadAll(part)
				if err != nil {
					log.Printf("Warning: could not read content of multipart part: %v", err)
					continue
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
