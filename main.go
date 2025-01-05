package main

import (
	"fmt"
	"log"
	"net/http"
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

		log.Println("Serving test.txt content")
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", "attachment; filename=test.txt")
		fmt.Fprint(w, testData)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}