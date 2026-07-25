package urls

import (
	"bytes"
	"github.com/arran4/header"
	"io"
	"net/http"
	"testing"
)

type mockClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestURLRule(t *testing.T) {
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == "https://example.com/image.jpg" {
				return &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Content-Type": []string{"image/jpeg"},
					},
					Body: io.NopCloser(bytes.NewBufferString("")),
				}, nil
			}
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	rule := urlCheckRule{}

	info := &header.PageInfo{
		OGImage: "https://example.com/image.jpg",
	}

	results := rule.Check(info, client)
	if len(results) != 0 {
		t.Errorf("Expected 0 errors for valid URL, got %d", len(results))
	}

	info = &header.PageInfo{
		OGImage: "https://example.com/notfound.jpg",
	}

	results = rule.Check(info, client)
	if len(results) != 1 {
		t.Errorf("Expected 1 error for not found URL, got %d", len(results))
	}
}
