package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kudzu-cms/kudzu/system/admin/user"
	"github.com/kudzu-cms/kudzu/system/db"
	"github.com/kudzu-cms/kudzu/system/item"
	"github.com/nilslice/jwt"
)

// ErrNoAuth should be used to report failed auth requests
var ErrNoAuth = errors.New("Auth failed for request")

// deprecating from API, but going to provide code here in case someone wants it
func typesHandler(res http.ResponseWriter, req *http.Request) {
	var types = []string{}
	for t, fn := range item.Types {
		if !hide(res, req, fn()) {
			types = append(types, t)
		}
	}

	j, err := toJSON(types)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, req, j)
}

func initHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if db.SystemInitComplete() {
		res.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	successJSON, _ := json.Marshal(map[string]interface{}{
		"success": true,
	})

	failureJSON, _ := json.Marshal(map[string]interface{}{
		"success": false,
	})

	err := req.ParseForm()
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(failureJSON)
		return
	}

	// get the site name from post to encode and use as secret
	name := []byte(req.FormValue("name") + db.NewEtag())
	secret := base64.StdEncoding.EncodeToString(name)
	req.Form.Set("client_secret", secret)

	// generate an Etag to use for response caching
	etag := db.NewEtag()
	req.Form.Set("etag", etag)

	// create and save admin user
	email := strings.ToLower(req.FormValue("email"))
	password := req.FormValue("password")
	usr, err := user.New(email, password)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(failureJSON)
		return
	}

	_, err = db.SetUser(usr)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(failureJSON)
		return
	}

	// set HTTP port which should be previously added to config cache
	port := db.ConfigCache("http_port").(string)
	req.Form.Set("http_port", port)

	// set initial user email as admin_email and make config
	req.Form.Set("admin_email", email)
	err = db.SetConfig(req.Form)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(failureJSON)
		return
	}

	// add _token cookie for login persistence
	week := time.Now().Add(time.Hour * 24 * 7)
	claims := map[string]interface{}{
		"exp":  week.Unix(),
		"user": usr.Email,
	}

	jwt.Secret([]byte(secret))
	token, err := jwt.New(claims)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(failureJSON)
		return
	}

	http.SetCookie(res, &http.Cookie{
		Name:    "_token",
		Value:   token,
		Expires: week,
		Path:    "/",
	})

	res.Write(successJSON)
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
	if !db.SystemInitComplete() {
		res.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	success := map[string]interface{}{
		"success": true,
	}
	successJSON, _ := json.Marshal(success)

	failure := map[string]interface{}{
		"success": false,
	}
	failureJSON, _ := json.Marshal(failure)

	switch req.Method {

	case http.MethodGet:
		res.Header().Set("Content-Type", "application/json")
		if user.IsValid(req) {
			res.Write(successJSON)
			return
		}

		res.WriteHeader(http.StatusUnauthorized)
		res.Write(failureJSON)
		return

	case http.MethodPost:
		res.Header().Set("Content-Type", "application/json")
		if user.IsValid(req) {
			res.Write(successJSON)
			return
		}

		err := req.ParseForm()
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(failureJSON)
			return
		}

		// check email & password
		j, err := db.User(strings.ToLower(req.FormValue("email")))
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusUnauthorized)
			res.Write(failureJSON)
			return
		}

		if j == nil {
			res.WriteHeader(http.StatusUnauthorized)
			res.Write(failureJSON)
			return
		}

		usr := &user.User{}
		err = json.Unmarshal(j, usr)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(failureJSON)
			return
		}

		if !user.IsUser(usr, req.FormValue("password")) {
			res.WriteHeader(http.StatusUnauthorized)
			res.Write(failureJSON)
			return
		}
		// create new token
		week := time.Now().Add(time.Hour * 24 * 7)
		claims := map[string]interface{}{
			"exp":  week,
			"user": usr.Email,
		}
		token, err := jwt.New(claims)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(failureJSON)
			return
		}

		// add it to cookie +1 week expiration
		http.SetCookie(res, &http.Cookie{
			Name:    "_token",
			Value:   token,
			Expires: week,
			Path:    "/",
		})
		res.Write(successJSON)
		return
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func contentsHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	it, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if hide(res, req, it()) {
		return
	}

	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if q.Get("count") == "" {
			count = 10
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	order := strings.ToLower(q.Get("order")) // string: sort order of posts by timestamp ASC / DESC (DESC default)
	if order != "asc" {
		order = "desc"
	}

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}

	_, bb := db.Query(t+"__sorted", opts)
	var result = []json.RawMessage{}
	for i := range bb {
		result = append(result, bb[i])
	}

	j, err := fmtJSON(result...)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, it(), j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// assert hookable
	get := it()
	hook, ok := get.(item.Hookable)
	if !ok {
		log.Println("[Response] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// hook before response
	j, err = hook.BeforeAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling BeforeAPIResponse:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, req, j)

	// hook after response
	err = hook.AfterAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling AfterAPIResponse:", err)
		return
	}
}

func contentHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	id := q.Get("id")
	t := q.Get("type")
	slug := q.Get("slug")

	if slug != "" {
		contentHandlerBySlug(res, req)
		return
	}

	if t == "" || id == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	pt, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	post, err := db.Content(t + ":" + id)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	p := pt()
	err = json.Unmarshal(post, p)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if hide(res, req, p) {
		return
	}

	push(res, req, p, post)

	j, err := fmtJSON(json.RawMessage(post))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, p, j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// assert hookable
	get := p
	hook, ok := get.(item.Hookable)
	if !ok {
		log.Println("[Response] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// hook before response
	j, err = hook.BeforeAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling BeforeAPIResponse:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, req, j)

	// hook after response
	err = hook.AfterAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling AfterAPIResponse:", err)
		return
	}
}

func contentHandlerBySlug(res http.ResponseWriter, req *http.Request) {
	slug := req.URL.Query().Get("slug")

	if slug == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// lookup type:id by slug key in __contentIndex
	t, post, err := db.ContentBySlug(slug)
	if err != nil {
		log.Println("Error finding content by slug:", slug, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	it, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	p := it()
	err = json.Unmarshal(post, p)
	if err != nil {
		log.Println(err)
		return
	}

	if hide(res, req, p) {
		return
	}

	push(res, req, p, post)

	j, err := fmtJSON(json.RawMessage(post))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, p, j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// assert hookable
	get := p
	hook, ok := get.(item.Hookable)
	if !ok {
		log.Println("[Response] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// hook before response
	j, err = hook.BeforeAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling BeforeAPIResponse:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, req, j)

	// hook after response
	err = hook.AfterAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling AfterAPIResponse:", err)
		return
	}

}

func uploadsHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	slug := req.URL.Query().Get("slug")
	if slug == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	upload, err := db.UploadBySlug(slug)
	if err != nil {
		log.Println("Error finding upload by slug:", slug, err)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	it := func() interface{} {
		return new(item.FileUpload)
	}

	push(res, req, it(), upload)

	j, err := fmtJSON(json.RawMessage(upload))
	if err != nil {
		log.Println("Error fmtJSON on upload:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, it(), j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, req, j)
}
