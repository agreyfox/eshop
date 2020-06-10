package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type News struct {
	item.Item

	Title    string `json:"title"`
	Text     string `json:"text"`
	Image    string `json:"image,omitempty"`
	Desc     string `json:"desc,description"`
	Number   int    `json:"number"`
	From     string `json:"from,omitemptyy"`
	Class    string `json:"class,omitempty"`
	Homepage bool   `json:"inhomepage,omitempty"`
}

// MarshalEditor writes a buffer of html to edit a News within the CMS
// and implements editor.Editable
func (n *News) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(n,
		// Take note that the first argument to these Input-like functions
		// is the string version of each News field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Title", n, map[string]string{
				"label":       "Title",
				"type":        "text",
				"placeholder": "Enter the Title here",
			}),
		},
		editor.Field{
			View: editor.Input("Text", n, map[string]string{
				"label":       "Text",
				"type":        "text",
				"placeholder": "Enter the Text here",
			}),
		},
		editor.Field{
			View: editor.Input("Image", n, map[string]string{
				"label":       "Image",
				"type":        "text",
				"placeholder": "Enter the Image here",
			}),
		},
		editor.Field{
			View: editor.Richtext("Desc", n, map[string]string{
				"label":       "Desc",
				"placeholder": "Enter the Desc here",
			}),
		},
		editor.Field{
			View: editor.Input("Number", n, map[string]string{
				"label":       "Number",
				"type":        "text",
				"placeholder": "Enter the Number here",
			}),
		},
		editor.Field{
			View: editor.Input("From", n, map[string]string{
				"label":       "From",
				"type":        "text",
				"placeholder": "Enter the From here",
			}),
		},
		editor.Field{
			View: editor.Input("Class", n, map[string]string{
				"label":       "Class",
				"type":        "text",
				"placeholder": "Enter the Class here",
			}),
		},
		editor.Field{
			View: editor.Input("Homepage", n, map[string]string{
				"label":       "Homepage",
				"type":        "text",
				"placeholder": "Enter the Homepage here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render News editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["News"] = func() interface{} { return new(News) }
}

// String defines how a News is printed. Update it using more descriptive
// fields from the News struct type
func (n *News) String() string {
	return fmt.Sprintf("News: %s", n.UUID)
}

func (o *News) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"title": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      1},
		"from": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      2},
		"text": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      3},
		"image": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      4},
		"number": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      10},
		"class": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      8},
		"inhomepage": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{"Show in homepage"},
			Order:      20},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   250,
	}
}
