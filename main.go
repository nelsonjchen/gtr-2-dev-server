package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const testContent = "abcdefghijklmnopqrstuvwxyz0123456789"

func main() {
	// Generate test content
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString(testContent)
	}
	testData := sb.String()

	http.HandleFunc("/setup.html", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accessed /setup.html")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html>
<body>
  <h1>Cookie Setup</h1>
  <button onclick="setCookie()">Set Cookie</button>
  <button onclick="clearCookie()">Clear Cookie</button>
  <script>
    function setCookie() {
      document.cookie = "testcookie=valid; path=/";
      alert("Cookie set!");
    }
    function clearCookie() {
      document.cookie = "testcookie=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";
      alert("Cookie cleared!");
    }
  </script>
</body>
</html>`)
	})

	http.HandleFunc("/download/test.txt", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accessed /download/test.txt")
		cookie, err := r.Cookie("testcookie")
		if err != nil || cookie.Value != "valid" {
			log.Println("Redirecting due to missing/invalid cookie")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", "attachment; filename=test.txt")
		w.Header().Set("Accept-Ranges", "bytes")

		fileSize := int64(len(testData))
		rangeHeader := r.Header.Get("Range")

		if rangeHeader == "" {
			log.Println("Serving full test.txt content")
			w.Header().Set("Content-Length", fmt.Sprint(fileSize))
			fmt.Fprint(w, testData)
			return
		}

		// Parse range header
		rangeParts := strings.Split(rangeHeader, "=")
		if len(rangeParts) != 2 || rangeParts[0] != "bytes" {
			http.Error(w, "Invalid Range header", http.StatusBadRequest)
			return
		}

		rangeSpec := strings.Split(rangeParts[1], "-")
		start, err := strconv.ParseInt(rangeSpec[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid Range header", http.StatusBadRequest)
			return
		}

		var end int64
		if rangeSpec[1] == "" {
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(rangeSpec[1], 10, 64)
			if err != nil || end >= fileSize {
				end = fileSize - 1
			}
		}

		if start > end || start < 0 {
			http.Error(w, "Invalid Range header", http.StatusBadRequest)
			return
		}

		contentLength := end - start + 1
		w.Header().Set("Content-Length", fmt.Sprint(contentLength))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		w.WriteHeader(http.StatusPartialContent)

		log.Printf("Serving partial content: bytes %d-%d/%d", start, end, fileSize)
		fmt.Fprint(w, testData[start:end+1])
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}