// Package admin desrcibes the admin view containing references to
// various managers and editors
package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/kudzu-cms/kudzu/system/admin/user"
	"github.com/kudzu-cms/kudzu/system/api/analytics"
	"github.com/kudzu-cms/kudzu/system/db"
	"github.com/kudzu-cms/kudzu/system/item"
)

var templatePathBase = "themes/admin/material/templates"

type admin struct {
	Logo    string
	Types   map[string]func() interface{}
	Subview template.HTML
}

// Admin ...
func Admin(view []byte) (_ []byte, err error) {
	cfg, err := db.Config("name")
	if err != nil {
		return
	}

	if cfg == nil {
		cfg = []byte("")
	}

	a := admin{
		Logo:    string(cfg),
		Types:   item.Types,
		Subview: template.HTML(view),
	}

	buf := &bytes.Buffer{}
	html := getTemplate("admin.start.tmpl.html") + getTemplate("admin.main.tmpl.html") + getTemplate("admin.end.tmpl.html")
	tmpl := template.Must(template.New("admin").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return
	}

	return buf.Bytes(), nil
}

// Init ...
func Init() ([]byte, error) {
	html := getTemplate("admin.start.tmpl.html") + getTemplate("admin.init.tmpl.html") + getTemplate("admin.end.tmpl.html")

	name, err := db.Config("name")
	if err != nil {
		return nil, err
	}

	if name == nil {
		name = []byte("")
	}

	a := admin{
		Logo: string(name),
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("init").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Login ...
func Login() ([]byte, error) {
	html := getTemplate("admin.start.tmpl.html") + getTemplate("admin.login.tmpl.html") + getTemplate("admin.end.tmpl.html")

	cfg, err := db.Config("name")
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		cfg = []byte("")
	}

	a := admin{
		Logo: string(cfg),
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("login").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ForgotPassword ...
func ForgotPassword() ([]byte, error) {
	html := getTemplate("admin.start.tmpl.html") + getTemplate("admin.forgot-password.tmpl.html") + getTemplate("admin.end.tmpl.html")

	cfg, err := db.Config("name")
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		cfg = []byte("")
	}

	a := admin{
		Logo: string(cfg),
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("forgotPassword").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RecoveryKey ...
func RecoveryKey() ([]byte, error) {
	html := getTemplate("admin.start.tmpl.html") + getTemplate("admin.recovery-key.tmpl.html") + getTemplate("admin.end.tmpl.html")

	cfg, err := db.Config("name")
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		cfg = []byte("")
	}

	a := admin{
		Logo: string(cfg),
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("recoveryKey").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UsersList ...
func UsersList(req *http.Request) ([]byte, error) {
	// get current user out to pass as data to execute template
	j, err := db.CurrentUser(req)
	if err != nil {
		return nil, err
	}

	var usr user.User
	err = json.Unmarshal(j, &usr)
	if err != nil {
		return nil, err
	}

	// get all users to list
	jj, err := db.UserAll()
	if err != nil {
		return nil, err
	}

	var usrs []user.User
	for i := range jj {
		var u user.User
		err = json.Unmarshal(jj[i], &u)
		if err != nil {
			return nil, err
		}
		if u.Email != usr.Email {
			usrs = append(usrs, u)
		}
	}

	// make buffer to execute html into then pass buffer's bytes to Admin
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("users").Parse(getTemplate("admin.user-list.tmpl.html")))
	data := map[string]interface{}{
		"User":  usr,
		"Users": usrs,
	}

	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	return Admin(buf.Bytes())
}

// Dashboard returns the admin view with analytics dashboard
func Dashboard() ([]byte, error) {
	buf := &bytes.Buffer{}
	data, err := analytics.ChartData()
	if err != nil {
		return nil, err
	}

	tmpl := template.Must(template.New("analytics").Parse(getTemplate("admin.analytics.tmpl.html")))
	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return Admin(buf.Bytes())
}

// Error400 creates a subview for a 400 error page
func Error400() ([]byte, error) {
	return Admin([]byte(getTemplate("error.400.tmpl.html")))
}

// Error404 creates a subview for a 404 error page
func Error404() ([]byte, error) {
	return Admin([]byte(getTemplate("error.404.tmpl.html")))
}

// Error405 creates a subview for a 405 error page
func Error405() ([]byte, error) {
	return Admin([]byte(getTemplate("error.405.tmpl.html")))
}

// Error500 creates a subview for a 500 error page
func Error500() ([]byte, error) {
	return Admin([]byte(getTemplate("error.500.tmpl.html")))
}

// ErrorMessage is a generic error message container, similar to Error500() and
// others in this package, ecxept it expects the caller to provide a title and
// message to describe to a view why the error is being shown
func ErrorMessage(title, message string) ([]byte, error) {
	eHTML := fmt.Sprintf(getTemplate("error.message.tmpl.html"), title, message)
	return Admin([]byte(eHTML))
}

func getTemplate(tmpl string) string {
	templatePath := templatePathBase + "/" + tmpl
	html, err := ioutil.ReadFile(templatePath)
	if err != nil {
		panic(err)
	}
	return string(html)
}
