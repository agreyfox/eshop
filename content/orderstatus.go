package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Orderstatus struct {
	item.Item

	Name  string `json:"name"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
	Desc  string `json:"desc"`
}

// MarshalEditor writes a buffer of html to edit a Orderstatus within the CMS
// and implements editor.Editable
func (o *Orderstatus) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(o,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Orderstatus field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", o, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Label", o, map[string]string{
				"label":       "Label",
				"type":        "text",
				"placeholder": "Enter the Label here",
			}),
		},
		editor.Field{
			View: editor.Input("Icon", o, map[string]string{
				"label":       "Icon",
				"type":        "text",
				"placeholder": "Enter the Icon here",
			}),
		},
		editor.Field{
			View: editor.Richtext("Desc", o, map[string]string{
				"label":       "Desc",
				"placeholder": "Enter the Desc here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Orderstatus editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Orderstatus"] = func() interface{} { return new(Orderstatus) }
}

// String defines how a Orderstatus is printed. Update it using more descriptive
// fields from the Orderstatus struct type
func (o *Orderstatus) String() string {
	return fmt.Sprintf("Orderstatus: %s", o.UUID)
}

func (o *Orderstatus) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"name": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      1},
		"label": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      2},
		"icon": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      6},

		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      7},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   200,
	}
}
