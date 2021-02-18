package editor

import (
	"bytes"
	"html"
	"strings"
	"text/template"

	"github.com/kudzu-cms/kudzu/system/theme"
)

// Input returns the []byte of an <input> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
// 	type Person struct {
//		item.Item
// 		editor editor.Editor
//
// 		Name string `json:"name"`
//		//...
// 	}
//
// 	func (p *Person) MarshalEditor() ([]byte, error) {
// 		view, err := editor.Form(p,
// 			editor.Field{
// 				View: editor.Input("Name", p, map[string]string{
// 					"label":       "Name",
// 					"type":        "text",
// 					"placeholder": "Enter the Name here",
// 				}),
// 			}
// 		)
// 	}
func Input(fieldName string, p interface{}, attrs map[string]string) []byte {
	e := NewElement("input", attrs["label"], fieldName, p, attrs)

	return DOMElementSelfClose(e)
}

// Textarea returns the []byte of a <textarea> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Textarea(fieldName string, p interface{}, attrs map[string]string) []byte {
	// add materialize css class to make UI correct
	className := "materialize-textarea"
	if _, ok := attrs["class"]; ok {
		class := attrs["class"]
		attrs["class"] = class + " " + className
	} else {
		attrs["class"] = className
	}

	e := NewElement("textarea", attrs["label"], fieldName, p, attrs)

	return DOMElement(e)
}

// Timestamp returns the []byte of an <input> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Timestamp(fieldName string, p interface{}, attrs map[string]string) []byte {
	var data string
	val := ValueFromStructField(fieldName, p)
	if val == "0" {
		data = ""
	} else {
		data = val
	}

	e := &Element{
		TagName: "input",
		Attrs:   attrs,
		Name:    TagNameFromStructField(fieldName, p),
		Label:   attrs["label"],
		Data:    data,
		ViewBuf: &bytes.Buffer{},
	}

	return DOMElementSelfClose(e)
}

type file struct {
	Name  string
	Value string
	Label string
}

// File returns the []byte of a <input type="file"> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func File(fieldName string, p interface{}, attrs map[string]string) []byte {
	name := TagNameFromStructField(fieldName, p)
	value := ValueFromStructField(fieldName, p)

	values := file{
		Name:  name,
		Value: value,
		Label: attrs["label"],
	}

	buf := &bytes.Buffer{}
	html := theme.LoadTemplateFromFilesystem("elements/file.tmpl.html")
	tmpl := template.Must(template.New("file").Parse(html))
	err := tmpl.Execute(buf, values)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// Richtext returns the []byte of a rich text editor (provided by http://summernote.org/) with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Richtext(fieldName string, p interface{}, attrs map[string]string) []byte {
	// create wrapper for richtext editor, which isolates the editor's css
	iso := []byte(`<div class="iso-texteditor input-field col s12"><label>` + attrs["label"] + `</label>`)
	isoClose := []byte(`</div>`)

	if _, ok := attrs["class"]; ok {
		attrs["class"] += "richtext " + fieldName
	} else {
		attrs["class"] = "richtext " + fieldName
	}

	if _, ok := attrs["id"]; ok {
		attrs["id"] += "richtext-" + fieldName
	} else {
		attrs["id"] = "richtext-" + fieldName
	}

	// create the target element for the editor to attach itself
	div := &Element{
		TagName: "div",
		Attrs:   attrs,
		Name:    "",
		Label:   "",
		Data:    "",
		ViewBuf: &bytes.Buffer{},
	}

	// create a hidden input to store the value from the struct
	val := ValueFromStructField(fieldName, p)
	name := TagNameFromStructField(fieldName, p)
	input := `<input type="hidden" name="` + name + `" class="richtext-value ` + fieldName + `" value="` + html.EscapeString(val) + `"/>`

	// build the dom tree for the entire richtext component
	iso = append(iso, DOMElement(div)...)
	iso = append(iso, []byte(input)...)
	iso = append(iso, isoClose...)

	script := `
	<script>
		$(function() {
			var _editor = $('.richtext.` + fieldName + `');
			var hidden = $('.richtext-value.` + fieldName + `');

			_editor.materialnote({
				height: 250,
				placeholder: '` + attrs["placeholder"] + `',
				toolbar: [
					['style', ['style']],
					['font', ['bold', 'italic', 'underline', 'clear', 'strikethrough', 'superscript', 'subscript']],
					['fontsize', ['fontsize']],
					['color', ['color']],
					['insert', ['link', 'picture', 'video', 'hr']],
					['para', ['ul', 'ol', 'paragraph']],
					['table', ['table']],
					['height', ['height']],
					['misc', ['codeview']]
				],
				// intercept file insertion, upload and insert img with new src
				onImageUpload: function(files) {
					var data = new FormData();
					data.append("file", files[0]);
					$.ajax({
						data: data,
						type: 'PUT',
						url: '/admin/edit/upload',
						cache: false,
						contentType: false,
						processData: false,
						success: function(resp) {
							var img = document.createElement('img');
							img.setAttribute('src', resp.data[0].url);
							_editor.materialnote('insertNode', img);
						},
						error: function(xhr, status, err) {
							console.log(status, err);
						}
					})

				}
			});

			// inject content into editor
			if (hidden.val() !== "") {
				_editor.code(hidden.val());
			}

			// update hidden input with encoded value on different events
			_editor.on('materialnote.change', function(e, content, $editable) {
				hidden.val(replaceBadChars(content));
			});

			_editor.on('materialnote.paste', function(e) {
				hidden.val(replaceBadChars(_editor.code()));
			});

			// bit of a hack to stop the editor buttons from causing a refresh when clicked
			$('.note-toolbar').find('button, i, a').on('click', function(e) { e.preventDefault(); });
		});
	</script>`

	return append(iso, []byte(script)...)
}

// Select returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Select(fieldName string, p interface{}, attrs, options map[string]string) []byte {
	// options are the value attr and the display value, i.e.
	// <option value="{map key}">{map value}</option>

	// find the field value in p to determine if an option is pre-selected
	fieldVal := ValueFromStructField(fieldName, p)

	if _, ok := attrs["class"]; ok {
		attrs["class"] += " browser-default"
	} else {
		attrs["class"] = "browser-default"
	}

	sel := NewElement("select", attrs["label"], fieldName, p, attrs)
	var opts []*Element

	// provide a call to action for the select element
	cta := &Element{
		TagName: "option",
		Attrs:   map[string]string{"disabled": "true", "selected": "true"},
		Data:    "Select an option...",
		ViewBuf: &bytes.Buffer{},
	}

	// provide a selection reset (will store empty string in db)
	reset := &Element{
		TagName: "option",
		Attrs:   map[string]string{"value": ""},
		Data:    "None",
		ViewBuf: &bytes.Buffer{},
	}

	opts = append(opts, cta, reset)

	for k, v := range options {
		optAttrs := map[string]string{"value": k}
		if k == fieldVal {
			optAttrs["selected"] = "true"
		}
		opt := &Element{
			TagName: "option",
			Attrs:   optAttrs,
			Data:    v,
			ViewBuf: &bytes.Buffer{},
		}

		opts = append(opts, opt)
	}

	return DOMElementWithChildrenSelect(sel, opts)
}

// Checkbox returns the []byte of a set of <input type="checkbox"> HTML elements
// wrapped in a <div> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Checkbox(fieldName string, p interface{}, attrs, options map[string]string) []byte {
	if _, ok := attrs["class"]; ok {
		attrs["class"] += "input-field col s12"
	} else {
		attrs["class"] = "input-field col s12"
	}

	div := NewElement("div", attrs["label"], fieldName, p, attrs)

	var opts []*Element

	// get the pre-checked options if this is already an existing post
	checkedVals := ValueFromStructField(fieldName, p)
	checked := strings.Split(checkedVals, "__kudzu")

	i := 0
	for k, v := range options {
		inputAttrs := map[string]string{
			"type":  "checkbox",
			"value": k,
			"id":    strings.Join(strings.Split(v, " "), "-"),
		}

		// check if k is in the pre-checked values and set to checked
		for _, x := range checked {
			if k == x {
				inputAttrs["checked"] = "checked"
			}
		}

		// create a *element manually using the modified TagNameFromStructFieldMulti
		// func since this is for a multi-value name
		input := &Element{
			TagName: "input",
			Attrs:   inputAttrs,
			Name:    TagNameFromStructFieldMulti(fieldName, i, p),
			Label:   v,
			Data:    "",
			ViewBuf: &bytes.Buffer{},
		}

		opts = append(opts, input)
		i++
	}

	return DOMElementWithChildrenCheckbox(div, opts)
}

type tag struct {
	Name  string
	Value string
}

type tags struct {
	GroupName  string
	GroupLabel string
	Items      []tag
}

// Tags returns the []byte of a tag input (in the style of Materialze 'Chips') with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
// @fixme Tags not currently being reloaded properly.
func Tags(fieldName string, p interface{}, attrs map[string]string) []byte {
	name := TagNameFromStructField(fieldName, p)

	// get the saved tags if this is already an existing post
	values := ValueFromStructField(fieldName, p)
	var tagItems []string
	if strings.Contains(values, "__kudzu") {
		tagItems = strings.Split(values, "__kudzu")
	}

	// case where there is only one tag stored, thus has no separator
	if len(values) > 0 && !strings.Contains(values, "__kudzu") {
		tagItems = append(tagItems, values)
	}

	var items []tag
	i := 0
	for _, tagValue := range tagItems {
		tagName := TagNameFromStructFieldMulti(fieldName, i, p)
		item := tag{Value: tagValue, Name: tagName}
		items = append(items, item)
		i++
	}
	group := tags{
		GroupName:  name,
		GroupLabel: attrs["label"],
		Items:      items,
	}

	buf := &bytes.Buffer{}
	html := theme.LoadTemplateFromFilesystem("elements/tags.tmpl.html")
	tmpl := template.Must(template.New("tags").Parse(html))
	err := tmpl.Execute(buf, group)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}
