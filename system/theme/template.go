package theme

import "io/ioutil"

var templatePathBase = "themes/admin/material/templates"

// LoadTemplateFromFilesystem retrieves a template from the filesystem
func LoadTemplateFromFilesystem(tmpl string) string {
	templatePath := templatePathBase + "/" + tmpl
	html, err := ioutil.ReadFile(templatePath)
	if err != nil {
		panic(err)
	}
	return string(html)
}
