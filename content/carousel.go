package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Carousel struct {
	item.Item

	Name   string `json:"name"`
	Alt    string `json:"alt,omitempty"`
	Game   string `json:"game,omitempty"`
	Image  string `json:"image,omitempty"`
	Desc   string `json:"description,omitempty"`
	Number int    `json:"number,omitempty"`
}

// MarshalEditor writes a buffer of html to edit a Carousel within the CMS
// and implements editor.Editable
func (c *Carousel) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(c,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Carousel field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", c, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Alt", c, map[string]string{
				"label":       "Alt",
				"type":        "text",
				"placeholder": "Enter the Alt here",
			}),
		},
		editor.Field{
			View: editor.Input("Image", c, map[string]string{
				"label":       "Image",
				"type":        "text",
				"placeholder": "Enter the Image here",
			}),
		},
		editor.Field{
			View: editor.Richtext("Desc", c, map[string]string{
				"label":       "Desc",
				"placeholder": "Enter the Desc here",
			}),
		},
		editor.Field{
			View: editor.Input("Number", c, map[string]string{
				"label":       "Number",
				"type":        "text",
				"placeholder": "Enter the Number here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Carousel editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Carousel"] = func() interface{} { return new(Carousel) }
}

// String defines how a Carousel is printed. Update it using more descriptive
// fields from the Carousel struct type
func (c *Carousel) String() string {
	return fmt.Sprintf("Carousel: %s", c.UUID)
}

func (o *Carousel) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"name": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      1},
		"alt": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      2},
		"game": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Game"},
			Required:   true,
			Order:      3},
		"image": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      4},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      20},
		"number": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      5},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   205,
	}
}
