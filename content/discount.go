package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Discount struct {
	item.Item

	Name      string `json:"name"`
	List      string `json:"list"`
	Currency  string `json:"currency"`
	SellText  string `json:"selltext"` //encourge selling
	Starttime string `json:"starttime"`
	Endtime   string `json:"endtime"`
	Online    bool   `json:"online"`
	Desc      string `json:"desc,omitempty"`
}

// MarshalEditor writes a buffer of html to edit a Discount within the CMS
// and implements editor.Editable
func (d *Discount) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(d,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Discount field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", d, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.InputRepeater("List", d, map[string]string{
				"label":       "List",
				"type":        "text",
				"placeholder": "Enter the List here",
			}),
		},
		editor.Field{
			View: editor.Input("Starttime", d, map[string]string{
				"label":       "Starttime",
				"type":        "text",
				"placeholder": "Enter the Starttime here",
			}),
		},
		editor.Field{
			View: editor.Input("Endtime", d, map[string]string{
				"label":       "Endtime",
				"type":        "text",
				"placeholder": "Enter the Endtime here",
			}),
		},
		editor.Field{
			View: editor.Input("Online", d, map[string]string{
				"label":       "Online",
				"type":        "text",
				"placeholder": "Enter the Online here",
			}),
		},
		editor.Field{
			View: editor.Richtext("Desc", d, map[string]string{
				"label":       "Desc",
				"placeholder": "Enter the Desc here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Discount editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Discount"] = func() interface{} { return new(Discount) }
}

// String defines how a Discount is printed. Update it using more descriptive
// fields from the Discount struct type
func (d *Discount) String() string {
	return fmt.Sprintf("Discount: %s", d.UUID)
}

/*
func (o *Discount) ContentStruct() map[string]item.FieldDescription {
	dd := map[string]item.FieldDescription{

		"name": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      1},
		"list": {
			Type:       "input",
			DataType:   "content",
			DataSource: []string{},
			Order:      2},
		"online": {
			Type:       "bool",
			DataType:   "content",
			DataSource: []string{},
			Order:      3},

		"starttime": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      6},
		"endtime": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      7},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      9},
	}
	//retStr, _ := json.Marshal(dd)
	return dd
}
*/
