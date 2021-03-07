package api

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/kudzu-cms/kudzu/system/admin/user"
	"github.com/kudzu-cms/kudzu/system/db"
)

// sendPreflight is used to respond to a cross-origin "OPTIONS" request
func sendPreflight(res http.ResponseWriter) {
	res.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.WriteHeader(200)
	return
}

func responseWithCORS(res http.ResponseWriter, req *http.Request) (http.ResponseWriter, bool) {
	if db.ConfigCache("cors_disabled").(bool) == true {
		// check origin matches config domain
		domain := db.ConfigCache("domain").(string)
		origin := req.Header.Get("Origin")
		u, err := url.Parse(origin)
		if err != nil {
			log.Println("Error parsing URL from request Origin header:", origin)
			return res, false
		}

		// hack to get dev environments to bypass cors since u.Host (below) will
		// be empty, based on Go's url.Parse function
		if domain == "localhost" {
			domain = ""
		}
		origin = u.Host

		// currently, this will check for exact match. will need feedback to
		// determine if subdomains should be allowed or allow multiple domains
		// in config
		if origin == domain {
			// apply limited CORS headers and return
			res.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
			res.Header().Set("Access-Control-Allow-Origin", domain)
			return res, true
		}

		// disallow request
		res.WriteHeader(http.StatusForbidden)
		return res, false
	}

	// apply full CORS headers and return
	res.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
	res.Header().Set("Access-Control-Allow-Origin", "*")

	return res, true
}

// AuthRequest ...
func AuthRequest(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if !user.IsValid(req) {
			res.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(res, req)
	})
}

// AuthCORS ...
func AuthCORS(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Access-Control-Allow-Origin", os.Getenv("KUDZU_CORS_ORIGIN"))
		res.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(res, req)
	})
}

// CORS wraps a HandlerFunc to respond to OPTIONS requests properly
func CORS(next http.HandlerFunc) http.HandlerFunc {
	return db.CacheControl(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res, cors := responseWithCORS(res, req)
		if !cors {
			return
		}

		if req.Method == http.MethodOptions {
			sendPreflight(res)
			return
		}

		next.ServeHTTP(res, req)
	}))
}
