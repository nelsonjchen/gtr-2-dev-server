package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func init() {
	// Generate test content
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString(testContent)
	}
	testData = sb.String()
}

func TestSetupHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/setup.html", nil)
	w := httptest.NewRecorder()

	mux := setupHandlers()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected content type text/html, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Cookie Setup") {
		t.Error("Expected HTML to contain 'Cookie Setup'")
	}
}

func TestDownloadHandler_NoCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/download/test.txt", nil)
	w := httptest.NewRecorder()

	mux := setupHandlers()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != "/setup.html" {
		t.Errorf("Expected redirect to /setup.html, got %s", location)
	}
}

func TestDownloadHandler_FullDownload(t *testing.T) {
	req := httptest.NewRequest("GET", "/download/test.txt", nil)
	req.AddCookie(&http.Cookie{Name: "testcookie", Value: "valid"})
	w := httptest.NewRecorder()

	mux := setupHandlers()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		t.Error("Expected Content-Length header")
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "attachment; filename=test.txt" {
		t.Errorf("Expected Content-Disposition header, got %s", contentDisposition)
	}

	body := w.Body.String()
	if len(body) != len(testContent)*1000 {
		t.Errorf("Expected full content length %d, got %d", len(testContent)*1000, len(body))
	}
}

func TestDownloadHandler_RangeRequests(t *testing.T) {
	tests := []struct {
		name        string
		rangeHeader string
		start       int
		end         int
		status      int
	}{
		{"Valid single range", "bytes=0-99", 0, 99, http.StatusPartialContent},
		{"Valid open-ended range", "bytes=100-", 100, len(testContent)*1000 - 1, http.StatusPartialContent},
		{"Valid suffix range", "bytes=-100", len(testContent)*1000 - 100, len(testContent)*1000 - 1, http.StatusPartialContent},
		{"Invalid range (start > end)", "bytes=100-50", 0, 0, http.StatusBadRequest},
		{"Invalid range (negative start)", "bytes=-100-200", 0, 0, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/download/test.txt", nil)
			req.Header.Set("Range", tt.rangeHeader)
			req.AddCookie(&http.Cookie{Name: "testcookie", Value: "valid"})
			w := httptest.NewRecorder()

			mux := setupHandlers()
			mux.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, resp.StatusCode)
			}

			if tt.status == http.StatusPartialContent {
				contentRange := resp.Header.Get("Content-Range")
				expectedRange := fmt.Sprintf("bytes %d-%d/%d", tt.start, tt.end, len(testContent)*1000)
				if contentRange != expectedRange {
					t.Errorf("Expected Content-Range %s, got %s", expectedRange, contentRange)
				}

				body := w.Body.String()
				if len(body) != tt.end-tt.start+1 {
					t.Errorf("Expected content length %d, got %d", tt.end-tt.start+1, len(body))
				}
			}
		})
	}
}