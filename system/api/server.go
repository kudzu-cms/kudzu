// Package api sets the various API handlers which provide an HTTP interface to
// kudzu content, and include the types and interfaces to enable client-side
// interactivity with the system.
package api

import (
	"net/http"
)

// Run adds Handlers to default http listener for API
func Run() {

	http.HandleFunc("/api/user/login", Record(AuthCORS(CORS(loginHandler))))

	http.HandleFunc("/api/user/logout", Record(AuthCORS(CORS(AuthRequest(logoutHandler)))))

	http.HandleFunc("/api/contents", Record(CORS(Gzip(contentsHandler))))

	http.HandleFunc("/api/contents/meta", Record(AuthCORS(CORS(AuthRequest(Gzip(contentsMetaHandler))))))

	http.HandleFunc("/api/content", Record(CORS(Gzip(contentHandler))))

	http.HandleFunc("/api/content/create", Record(AuthCORS(CORS(AuthRequest(createContentHandler)))))

	http.HandleFunc("/api/content/update", Record(AuthCORS(CORS(AuthRequest(updateContentHandler)))))

	http.HandleFunc("/api/content/delete", Record(AuthCORS(CORS(AuthRequest(deleteContentHandler)))))

	http.HandleFunc("/api/search", Record(CORS(Gzip(searchContentHandler))))

	http.HandleFunc("/api/uploads", Record(CORS(Gzip(uploadsHandler))))

	http.HandleFunc("/api/uploads/delete", Record(AuthCORS(CORS(Gzip(uploadsDeleteHandler)))))

	http.HandleFunc("/api/system/init", Record(AuthCORS(CORS(initHandler))))
}
