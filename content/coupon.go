package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Coupon struct {
	item.Item

	Name          string  `json:"name"`
	Game          string  `json:"game"`
	Price         float64 `json:"price"`
	Type          string  `json:"type"`
	Currency      string  `json:"currency"`
	InitialAmount float64 `json:"initial_amount"`
	Starttime     string  `json:"starttime"`
	Endtime       string  `json:"endtime"`
	Code          string  `json:"code"`
	Desc          string  `json:"desc"`
	Meta          string  `json:"meta"`
}

// MarshalEditor writes a buffer of html to edit a Coupon within the CMS
// and implements editor.Editable
func (c *Coupon) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(c,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Coupon field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", c, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Game", c, map[string]string{
				"label":       "Game",
				"type":        "text",
				"placeholder": "Enter the Game here",
			}),
		},
		editor.Field{
			View: editor.Input("Price", c, map[string]string{
				"label":       "Price",
				"type":        "text",
				"placeholder": "Enter the Price here",
			}),
		},
		editor.Field{
			View: editor.Input("Type", c, map[string]string{
				"label":       "Type",
				"type":        "text",
				"placeholder": "Enter the Type here",
			}),
		},
		editor.Field{
			View: editor.Input("InitialAmount", c, map[string]string{
				"label":       "InitialAmount",
				"type":        "text",
				"placeholder": "Enter the InitialAmount here",
			}),
		},
		editor.Field{
			View: editor.Input("Starttime", c, map[string]string{
				"label":       "Starttime",
				"type":        "text",
				"placeholder": "Enter the Starttime here",
			}),
		},
		editor.Field{
			View: editor.Input("Endtime", c, map[string]string{
				"label":       "Endtime",
				"type":        "text",
				"placeholder": "Enter the Endtime here",
			}),
		},
		editor.Field{
			View: editor.Input("Code", c, map[string]string{
				"label":       "Code",
				"type":        "text",
				"placeholder": "Enter the Code here",
			}),
		},
		editor.Field{
			View: editor.Richtext("Desc", c, map[string]string{
				"label":       "Desc",
				"placeholder": "Enter the Desc here",
			}),
		},
		editor.Field{
			View: editor.Input("Meta", c, map[string]string{
				"label":       "Meta",
				"type":        "text",
				"placeholder": "Enter the Meta here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Coupon editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Coupon"] = func() interface{} { return new(Coupon) }
}

// String defines how a Coupon is printed. Update it using more descriptive
// fields from the Coupon struct type
func (c *Coupon) String() string {
	return fmt.Sprintf("Coupon: %s", c.UUID)
}

/*
func (o *Coupon) ContentStruct() map[string]item.FieldDescription {
	dd := map[string]item.FieldDescription{

		"name": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      1},
		"game": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{},
			Order:      2},
		"price": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      4},
		"type": {
			Type:       "select",
			DataType:   "field",
			DataSource: []string{},
			Order:      3},
		"initial_amount": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      5},
		"code": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      8},

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
		"meta": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      10},
	}
	//retStr, _ := json.Marshal(dd)
	return dd
}
*/
