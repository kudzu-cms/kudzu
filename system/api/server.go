// Package api sets the various API handlers which provide an HTTP interface to
// kudzu content, and include the types and interfaces to enable client-side
// interactivity with the system.
package api

import (
	"net/http"
)

// Run adds Handlers to default http listener for API
func Run() {

	http.HandleFunc("/api/user/login", Record(CORS(AuthCORS(loginHandler))))

	http.HandleFunc("/api/contents", Record(CORS(Gzip(contentsHandler))))

	http.HandleFunc("/api/contents/meta", Record(CORS(AuthCORS(AuthRequest(Gzip(contentsMetaHandler))))))

	http.HandleFunc("/api/content", Record(CORS(Gzip(contentHandler))))

	http.HandleFunc("/api/content/create", Record(CORS(AuthCORS(AuthRequest(createContentHandler)))))

	http.HandleFunc("/api/content/update", Record(CORS(AuthCORS(AuthRequest(updateContentHandler)))))

	http.HandleFunc("/api/content/delete", Record(CORS(AuthCORS(AuthRequest(deleteContentHandler)))))

	http.HandleFunc("/api/search", Record(CORS(Gzip(searchContentHandler))))

	http.HandleFunc("/api/uploads", Record(CORS(Gzip(uploadsHandler))))

	http.HandleFunc("/api/system/init", Record(CORS(AuthCORS(initHandler))))
}
