package email

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		rawEmail      string
		wantMessageID string
		wantFrom      string
		wantTo        string
		wantSubject   string
		wantBody      string
		wantURLs      []string
		wantErr       bool
	}{
		{
			name: "Plain Text Email",
			rawEmail: `From: sender@example.com
To: recipient@example.com
Subject: Test Subject
Message-ID: <plain@example.com>

This is the body. Check out http://example.com. 
Another link: https://google.com/search?q=test`,
			wantMessageID: "plain@example.com",
			wantFrom:      "<sender@example.com>",
			wantTo:        "<recipient@example.com>",
			wantSubject:   "Test Subject",
			wantBody:      "This is the body. Check out http://example.com. Another link: https://google.com/search?q=test",
			wantURLs:      []string{"http://example.com", "https://google.com/search?q=test"},
			wantErr:       false,
		},
		{
			name: "HTML Email",
			rawEmail: `From: "HTML Sender" <sender@example.com>
To: "HTML Recipient" <recipient@example.com>
Subject: HTML Test
Message-ID: <html@example.com>
Content-Type: text/html

<h1>Hello</h1><p>This is a <a href="https://example.org">link</a>.</p>`,
			wantMessageID: "html@example.com",
			wantFrom:      `"HTML Sender" <sender@example.com>`,
			wantTo:        `"HTML Recipient" <recipient@example.com>`,
			wantSubject:   "HTML Test",
			wantBody:      "Hello This is a link .",
			wantURLs:      []string{"https://example.org"},
			wantErr:       false,
		},
		{
			name: "Multipart Email",
			rawEmail: `From: multipart@example.com
To: recipient@example.com
Subject: Multipart Test
Message-ID: <multipart@example.com>
Content-Type: multipart/alternative; boundary=boundary

--boundary
Content-Type: text/plain; charset="utf-8"

Plain text part. URL: http://plain.com
--boundary
Content-Type: text/html; charset="utf-8"

<p>HTML part. URL: <a href="http://html.com">html</a></p>
--boundary--
`,
			wantMessageID: "multipart@example.com",
			wantFrom:      "<multipart@example.com>",
			wantTo:        "<recipient@example.com>",
			wantSubject:   "Multipart Test",
			wantBody:      "Plain text part. URL: http://plain.com\n HTML part. URL: html",
			wantURLs:      []string{"http://plain.com", "http://html.com"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace \n with \r\n for email line endings
			rawEmailWithCRLF := strings.ReplaceAll(tt.rawEmail, "\n", "\r\n")
			got, err := Parse(strings.NewReader(rawEmailWithCRLF))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.MessageID != tt.wantMessageID {
				t.Errorf("Parse() MessageID = %v, want %v", got.MessageID, tt.wantMessageID)
			}
			if len(got.From) > 0 && got.From[0].String() != tt.wantFrom {
				t.Errorf("Parse() From = %v, want %v", got.From[0].String(), tt.wantFrom)
			}
			if len(got.To) > 0 && got.To[0].String() != tt.wantTo {
				t.Errorf("Parse() To = %v, want %v", got.To[0].String(), tt.wantTo)
			}
			if got.Subject != tt.wantSubject {
				t.Errorf("Parse() Subject = %v, want %v", got.Subject, tt.wantSubject)
			}

			// Sort slices for consistent comparison
			sort.Strings(got.URLs)
			sort.Strings(tt.wantURLs)
			if !reflect.DeepEqual(got.URLs, tt.wantURLs) {
				t.Errorf("Parse() URLs = %v, want %v", got.URLs, tt.wantURLs)
			}

			// Normalize whitespace for body comparison
			normalize := func(s string) string {
				return strings.Join(strings.Fields(s), " ")
			}
			if normalize(got.Body) != normalize(tt.wantBody) {
				t.Errorf("Parse() Body = \"%v\", want \"%v\"", normalize(got.Body), normalize(tt.wantBody))
			}
		})
	}
}

func TestExtractBodyAndURLs_Deduplication(t *testing.T) {
	rawEmail := `From: dedupe@example.com
To: recipient@example.com
Subject: Deduplication Test
Message-ID: <dedupe@example.com>

URL 1: http://example.com. URL 2: http://example.com. URL 3: <a href="http://example.com">example</a>`
	rawEmailWithCRLF := strings.ReplaceAll(rawEmail, "\n", "\r\n")
	parsed, err := Parse(strings.NewReader(rawEmailWithCRLF))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	wantURLs := []string{"http://example.com"}
	if !reflect.DeepEqual(parsed.URLs, wantURLs) {
		t.Errorf("Expected URLs to be deduplicated. got %v, want %v", parsed.URLs, wantURLs)
	}
}

func TestExtractBodyAndURLs_URLTrimming(t *testing.T) {
	rawEmail := `From: trim@example.com
To: recipient@example.com
Subject: Trimming Test
Message-ID: <trim@example.com>

Check this out: http://example.com/page, or this one: http://google.com).`
	rawEmailWithCRLF := strings.ReplaceAll(rawEmail, "\n", "\r\n")
	parsed, err := Parse(strings.NewReader(rawEmailWithCRLF))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	wantURLs := []string{"http://example.com/page", "http://google.com"}
	sort.Strings(parsed.URLs)
	sort.Strings(wantURLs)
	if !reflect.DeepEqual(parsed.URLs, wantURLs) {
		t.Errorf("Expected URLs to be trimmed. got %v, want %v", parsed.URLs, wantURLs)
	}
}