package template

import "io/ioutil"

var templatePathBase = "themes/admin/material/templates"

// LoadFromFilesystem retrieves a template from the filesystem
func LoadFromFilesystem(tmpl string) string {
	templatePath := templatePathBase + "/" + tmpl
	html, err := ioutil.ReadFile(templatePath)
	if err != nil {
		panic(err)
	}
	return string(html)
}
