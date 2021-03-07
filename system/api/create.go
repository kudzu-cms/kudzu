package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/kudzu-cms/kudzu/system/admin/upload"
	"github.com/kudzu-cms/kudzu/system/db"
	"github.com/kudzu-cms/kudzu/system/item"

	"github.com/gorilla/schema"
)

func contentsMetaHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Return type information if no query parameters are provided.
	//
	// @todo This needs to be implemented recursively to better handle nested
	// types.
	//
	// @todo It also needs to be determined whether or not nested types should
	// be flattened in the returned JSON. Probably, the answer is yes, but it
	// still needs to be considered whether or not this could result in field name
	// collisions. The item.Item type could be treated as special case that is
	// flattened, while others will not be.
	//
	// @todo This endpoint should return field types in their presentation format.
	// Field types shouldn't  be returned as their true internal data type.
	// For instance, `int64` could be returned simply as `int` and `uuid.UUID`
	// could be returned as `string` because external systems are welcome to think
	// of this type as being represented by a string.
	types := map[string]interface{}{}
	for name, t := range item.Types {
		tInst := t()
		tJson, _ := json.Marshal(tInst)
		typeSchema := map[string]interface{}{}
		err := json.Unmarshal(tJson, &typeSchema)
		if err != nil {
			panic(err)
		}
		ref := reflect.ValueOf(tInst).Elem()
		for i := 0; i < ref.NumField(); i++ {
			field := ref.Type().Field(i)
			fieldName := strings.TrimSuffix(strings.TrimPrefix(string(field.Tag), "json:\""), "\"")
			fieldType := field.Type.String()
			if fieldType != "item.Item" {
				typeSchema[fieldName] = fieldType
			} else {
				itemInst := item.Item{}
				itemRef := reflect.ValueOf(&itemInst).Elem()
				// To return in nested representation:
				// itemSchema := map[string]interface{}{}
				for j := 0; j < itemRef.NumField(); j++ {
					field := itemRef.Type().Field(j)
					fieldName := strings.TrimSuffix(strings.TrimPrefix(string(field.Tag), "json:\""), "\"")
					fieldType := field.Type.String()
					// To return in nested representation:
					// itemSchema[fieldName] = fieldType
					typeSchema[fieldName] = fieldType
				}
				// To return in nested representation:
				// typeSchema[fieldName] = itemSchema
			}
		}
		types[name] = typeSchema
	}
	jsonResponse, err := json.Marshal(interface{}(types))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.Write(jsonResponse)
	return
}

func createContentHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		log.Println("[Create] error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	t := req.URL.Query().Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	p, found := item.Types[t]
	if !found {
		log.Println("[Create] attempt to submit unknown type:", t, "from:", req.RemoteAddr)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	post := p()

	ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UnixNano()/int64(time.Millisecond))
	req.PostForm.Set("timestamp", ts)
	req.PostForm.Set("updated", ts)

	urlPaths, err := upload.StoreFiles(req)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	for name, urlPath := range urlPaths {
		req.PostForm.Set(name, urlPath)
	}

	// check for any multi-value fields (ex. checkbox fields)
	// and correctly format for db storage. Essentially, we need
	// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
	fieldOrderValue := make(map[string]map[string][]string)
	for k, v := range req.PostForm {
		if strings.Contains(k, ".") {
			fo := strings.Split(k, ".")

			// put the order and the field value into map
			field := string(fo[0])
			order := string(fo[1])
			if len(fieldOrderValue[field]) == 0 {
				fieldOrderValue[field] = make(map[string][]string)
			}

			// orderValue is 0:[?type=Thing&id=1]
			orderValue := fieldOrderValue[field]
			orderValue[order] = v
			fieldOrderValue[field] = orderValue

			// discard the post form value with name.N
			req.PostForm.Del(k)
		}

	}

	// add/set the key & value to the post form in order
	for f, ov := range fieldOrderValue {
		for i := 0; i < len(ov); i++ {
			position := fmt.Sprintf("%d", i)
			fieldValue := ov[position]

			if req.PostForm.Get(f) == "" {
				for i, fv := range fieldValue {
					if i == 0 {
						req.PostForm.Set(f, fv)
					} else {
						req.PostForm.Add(f, fv)
					}
				}
			} else {
				for _, fv := range fieldValue {
					req.PostForm.Add(f, fv)
				}
			}
		}
	}

	hook, ok := post.(item.Hookable)
	if !ok {
		log.Println("[Create] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Let's be nice and make a proper item for the Hookable methods
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, req.PostForm)
	if err != nil {
		log.Println("Error decoding post form for edit handler:", t, err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hook.BeforeAPICreate(res, req)
	if err != nil {
		log.Println("[Create] error calling BeforeCreate:", err)
		return
	}

	err = hook.BeforeSave(res, req)
	if err != nil {
		log.Println("[Create] error calling BeforeSave:", err)
		return
	}

	id, err := db.SetContent(t+":-1", req.PostForm)
	if err != nil {
		log.Println("[Create] error calling SetContent:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// set the target in the context so user can get saved value from db in hook
	ctx := context.WithValue(req.Context(), "target", fmt.Sprintf("%s:%d", t, id))
	req = req.WithContext(ctx)

	err = hook.AfterSave(res, req)
	if err != nil {
		log.Println("[Create] error calling AfterSave:", err)
		return
	}

	err = hook.AfterAPICreate(res, req)
	if err != nil {
		log.Println("[Create] error calling AfterAccept:", err)
		return
	}

	// create JSON response to send data back to client
	data := map[string]interface{}{
		"id":     id,
		"status": "public",
		"type":   t,
	}

	resp := map[string]interface{}{
		"data": []map[string]interface{}{
			data,
		},
	}

	j, err := json.Marshal(resp)
	if err != nil {
		log.Println("[Create] error marshalling response to JSON:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(j)
	if err != nil {
		log.Println("[Create] error writing response:", err)
		return
	}

}
