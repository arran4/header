package images

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"testing"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

// A tiny 1x1 GIF
const testGifBase64 = "R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"

// A tiny 2x2 GIF
const testGif2x2Base64 = "R0lGODlhAgACAPAAAP///wAAACwAAAAAAgACAAACAkQBADs="

type mockHTTPClient struct {
	responses map[string]*http.Response
	errs      map[string]error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if err, ok := m.errs[req.URL.String()]; ok {
		return nil, err
	}
	if resp, ok := m.responses[req.URL.String()]; ok {
		// Because the body is read in DecodeConfig, we need to create a new reader each time
		// so it can be re-read if multiple tests use the same URL concurrently or serially.
		bodyData, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewReader(bodyData))

		return &http.Response{
			StatusCode: resp.StatusCode,
			Header:     resp.Header,
			Body:       io.NopCloser(bytes.NewReader(bodyData)),
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}, nil
}

func makeGIFResponse(b64 string, contentType string) *http.Response {
	data, _ := base64.StdEncoding.DecodeString(b64)
	header := make(http.Header)
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
}

func TestImagesRule(t *testing.T) {
	rule := imagesRule{}

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			"http://example.com/1x1.gif":   makeGIFResponse(testGifBase64, "image/gif"),
			"http://example.com/2x2.gif":   makeGIFResponse(testGif2x2Base64, "image/gif"),
			"http://example.com/badct.gif": makeGIFResponse(testGifBase64, "text/html"),
			"http://example.com/noct.gif":  makeGIFResponse(testGifBase64, ""),
		},
		errs: map[string]error{},
	}

	tests := []struct {
		name     string
		info     *header.PageInfo
		expected []lint.Result
	}{
		{
			name: "Valid Image",
			info: &header.PageInfo{
				OGImage:       "http://example.com/1x1.gif",
				OGImageWidth:  "1",
				OGImageHeight: "1",
				OGImageType:   "image/gif",
			},
			expected: nil,
		},
		{
			name: "Mismatched Width",
			info: &header.PageInfo{
				OGImage:       "http://example.com/1x1.gif",
				OGImageWidth:  "2",
				OGImageHeight: "1",
				OGImageType:   "image/gif",
			},
			expected: []lint.Result{
				{Severity: lint.SeverityWarning, Message: "Image at http://example.com/1x1.gif width mismatch: metadata says 2, actual is 1"},
			},
		},
		{
			name: "Mismatched Height",
			info: &header.PageInfo{
				OGImage:       "http://example.com/1x1.gif",
				OGImageWidth:  "1",
				OGImageHeight: "2",
				OGImageType:   "image/gif",
			},
			expected: []lint.Result{
				{Severity: lint.SeverityWarning, Message: "Image at http://example.com/1x1.gif height mismatch: metadata says 2, actual is 1"},
			},
		},
		{
			name: "Bad Content Type",
			info: &header.PageInfo{
				OGImage: "http://example.com/badct.gif",
			},
			expected: []lint.Result{
				{Severity: lint.SeverityError, Message: "Image URL http://example.com/badct.gif returned non-image Content-Type: text/html"},
			},
		},
		{
			name: "Missing Content Type",
			info: &header.PageInfo{
				OGImage: "http://example.com/noct.gif",
			},
			expected: []lint.Result{
				{Severity: lint.SeverityWarning, Message: "Image URL http://example.com/noct.gif missing Content-Type header"},
			},
		},
		{
			name: "Mismatched Expected Content Type",
			info: &header.PageInfo{
				OGImage:     "http://example.com/1x1.gif",
				OGImageType: "image/png",
			},
			expected: []lint.Result{
				{Severity: lint.SeverityWarning, Message: "Image URL http://example.com/1x1.gif expected type image/png but got image/gif"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := rule.Check(tt.info, client)

			if len(results) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d. Results: %v", len(tt.expected), len(results), results)
			}

			for i, res := range results {
				if res.Severity != tt.expected[i].Severity || res.Message != tt.expected[i].Message {
					t.Errorf("expected result %d to be {Severity: %d, Message: %q}, got {Severity: %d, Message: %q}",
						i, tt.expected[i].Severity, tt.expected[i].Message, res.Severity, res.Message)
				}
			}
		})
	}
}
