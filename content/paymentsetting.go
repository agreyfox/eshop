package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type PaymentSetting struct {
	item.Item

	Name        string  `json:"name"`
	Value       float32 `json:"value"`
	ValueString string  `json:"valueString"`
	ValueInt    int     `json:"valueInt"`
	ValueBool   bool    `json;"valueBool"`
	Description string  `json:"description"`
	Status      bool    `json:"status"`
}

// MarshalEditor writes a buffer of html to edit a PaymentSetting within the CMS
// and implements editor.Editable
func (p *PaymentSetting) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(p,
		// Take note that the first argument to these Input-like functions
		// is the string version of each PaymentSetting field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", p, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Value", p, map[string]string{
				"label":       "Value",
				"type":        "text",
				"placeholder": "Enter the Value here",
			}),
		},
		editor.Field{
			View: editor.Input("Description", p, map[string]string{
				"label":       "Description",
				"type":        "text",
				"placeholder": "Enter the Description here",
			}),
		},
		editor.Field{
			View: editor.Input("Status", p, map[string]string{
				"label":       "Status",
				"type":        "text",
				"placeholder": "Enter the Status here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render PaymentSetting editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["PaymentSetting"] = func() interface{} { return new(PaymentSetting) }
}

// String defines how a PaymentSetting is printed. Update it using more descriptive
// fields from the PaymentSetting struct type
func (p *PaymentSetting) String() string {
	return fmt.Sprintf("PaymentSetting: %s", p.UUID)
}

func (o *PaymentSetting) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"name": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      1},
		"value": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Help:       "收费百分百",
			Order:      20},
		"valueString": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "字符串数值",
			Order:      30},
		"valueInt": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "整型数值",
			Order:      40},
		"valueBool": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "逻辑值",
			Order:      50},

		"status": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{},
			Order:      70},

		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      80},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   210,
	}
}
