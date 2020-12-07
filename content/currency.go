package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Currency struct {
	item.Item

	Name     string  `json:"name"`
	Symbol   string  `json:"symbol"`
	Rate     float64 `json:"rate"`
	Icon     string  `json:"icon,omitempty"`
	Paypal   string  `json:"paypal,omitempty"`
	SkrilL   string  `json:"skrill,omitempty"`
	Payssion string  `json:"payssion,omitempty"`
	Static   string  `json:"static,omitempty"`
	Desc     string  `json:"desc,omitempty"`
}

// MarshalEditor writes a buffer of html to edit a Currency within the CMS
// and implements editor.Editable
func (c *Currency) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(c,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Currency field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", c, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Symbol", c, map[string]string{
				"label":       "Symbol",
				"type":        "text",
				"placeholder": "Enter the Symbol here",
			}),
		},
		editor.Field{
			View: editor.Input("Rate", c, map[string]string{
				"label":       "Rate",
				"type":        "text",
				"placeholder": "Enter the Rate here",
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
		return nil, fmt.Errorf("Failed to render Currency editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Currency"] = func() interface{} { return new(Currency) }
}

// String defines how a Currency is printed. Update it using more descriptive
// fields from the Currency struct type
func (c *Currency) String() string {
	return fmt.Sprintf("Currency: %s", c.UUID)
}

func (o *Currency) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{

		"name": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      1},

		"symbol": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      2},
		"rate": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      3},
		"icon": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      4},
		"paypal": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "此货币在paypal中的表示方法",
			Order:      7},
		"payssion": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "此货币在payssion中的表示方法",
			Order:      8},
		"skrill": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "此货币在skrill中的表示方法",
			Order:      9},
		"static": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "此货币在static中的表示方法",
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
		"no":   60,
	}
}
