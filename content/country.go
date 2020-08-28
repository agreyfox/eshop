package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Country struct {
	item.Item

	Name     string `json:"name"`
	Fullname string `json:"fullname,omitempty"`
	Icon     string `json:"icon,omitempty"`
	Desc     string `json:"description:omitempty"`
	Paypal   string `json:"paypal,omitempty"`
	SkrilL   string `json:"skrill,omitempty"`
	Payssion string `json:"payssion,omitempty"`
}

// MarshalEditor writes a buffer of html to edit a Country within the CMS
// and implements editor.Editable
func (c *Country) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(c,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Country field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", c, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Fullname", c, map[string]string{
				"label":       "Fullname",
				"type":        "text",
				"placeholder": "Enter the Fullname here",
			}),
		},
		editor.Field{
			View: editor.Input("Icon", c, map[string]string{
				"label":       "Icon",
				"type":        "text",
				"placeholder": "Enter the Icon here",
			}),
		},
		editor.Field{
			View: editor.Richtext("Desc", c, map[string]string{
				"label":       "Desc",
				"placeholder": "Enter the Desc here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Country editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Country"] = func() interface{} { return new(Country) }
}

// String defines how a Country is printed. Update it using more descriptive
// fields from the Country struct type
func (c *Country) String() string {
	return fmt.Sprintf("Country: %s", c.UUID)
}

func (o *Country) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"name": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      1},
		"fullname": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      2},

		"icon": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      6},
		"paypal": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      7},
		"payssion": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      8},
		"skrill": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      9},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      10},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   230,
	}
}
