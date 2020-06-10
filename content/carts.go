package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Carts struct {
	item.Item

	Email    string `json:"email"`
	Items    string `json:"items"`
	Coupon   string `json:"coupon,omitempty"`
	Comments string `json:"comments,omitempty"`
}

// MarshalEditor writes a buffer of html to edit a Carts within the CMS
// and implements editor.Editable
func (c *Carts) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(c,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Carts field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Email", c, map[string]string{
				"label":       "Email",
				"type":        "text",
				"placeholder": "Enter the Email here",
			}),
		},
		editor.Field{
			View: editor.Input("Items", c, map[string]string{
				"label":       "Items",
				"type":        "text",
				"placeholder": "Enter the Items here",
			}),
		},
		editor.Field{
			View: editor.Input("Coupon", c, map[string]string{
				"label":       "Coupon",
				"type":        "text",
				"placeholder": "Enter the Coupon here",
			}),
		},
		editor.Field{
			View: editor.Input("Comments", c, map[string]string{
				"label":       "Comments",
				"type":        "text",
				"placeholder": "Enter the Comments here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Carts editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Carts"] = func() interface{} { return new(Carts) }
}

// String defines how a Carts is printed. Update it using more descriptive
// fields from the Carts struct type
func (c *Carts) String() string {
	return fmt.Sprintf("Carts: %s", c.UUID)
}

/*
func (o *Carts) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"title": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      1},
		"text": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      2},
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
		"namber": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      5},
		"from": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      3},
		"class": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      10},
		"inhomepage": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      6},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   210,
	}
}
*/
