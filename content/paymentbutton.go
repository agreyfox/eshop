package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Paymentbutton struct {
	item.Item

	Name       string  `json:"name"`
	ImgUrl     string  `json:"img_url"`
	ImgUrlBak  string  `json:"img_url2"`
	Note       string  `json:"note"`
	Payment    string  `json:"payment"`
	Channel    string  `json:"channel"`
	PaymentFee float64 `json:"payment_fee"`
	Order      int     `json:"order"`
	FeeType    int     `json:"fee_type"`
	Helper     string  `json:"helper"`
}

// MarshalEditor writes a buffer of html to edit a Paymentbutton within the CMS
// and implements editor.Editable
func (p *Paymentbutton) MarshalEditor() ([]byte, error) {
	view, _ := editor.Form(p,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Paymentbutton field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", p, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("ImgUrl", p, map[string]string{
				"label":       "ImgUrl",
				"type":        "text",
				"placeholder": "Enter the ImgUrl here",
			}),
		},
		editor.Field{
			View: editor.Input("Note", p, map[string]string{
				"label":       "Note",
				"type":        "text",
				"placeholder": "Enter the Note here",
			}),
		},
		editor.Field{
			View: editor.Input("Payment", p, map[string]string{
				"label":       "Payment",
				"type":        "text",
				"placeholder": "Enter the Payment here",
			}),
		},
		editor.Field{
			View: editor.Input("Channel", p, map[string]string{
				"label":       "Channel",
				"type":        "text",
				"placeholder": "Enter the Channel here",
			}),
		},

		editor.Field{
			View: editor.Input("Order", p, map[string]string{
				"label":       "Order",
				"type":        "text",
				"placeholder": "Enter the Order here",
			}),
		},
		editor.Field{
			View: editor.Input("FeeType", p, map[string]string{
				"label":       "FeeType",
				"type":        "text",
				"placeholder": "Enter the FeeType here",
			}),
		},
		editor.Field{
			View: editor.Input("Helper", p, map[string]string{
				"label":       "Helper",
				"type":        "text",
				"placeholder": "Enter the Helper here",
			}),
		},
	)

	return view, nil
}

func init() {
	item.Types["Paymentbutton"] = func() interface{} { return new(Paymentbutton) }
}

func (p *Paymentbutton) String() string {
	return fmt.Sprintf("Paymentbutton: %s", p.UUID)
}

func (o *Paymentbutton) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"name": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      1},
		"img_url": {
			Type:       "file",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Help:       "选择付款按钮使用的图片",
			Order:      20},
		"img_url2": {
			Type:       "input",
			DataType:   "field",
			Required:   false,
			DataSource: []string{},
			Help:       "填入按钮使用的图片,备用",
			Order:      21},
		"note": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "按钮下方的文字说明",
			Order:      30},
		"payment": {
			Type:       "select",
			DataType:   "field",
			Required:   true,
			DataSource: []string{"paypal", "payssion", "skrill", "static"},
			Help:       "支付方法选择:paypal,payssion,skrill,static",
			Order:      40},
		"channel": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Help:       "支付手段，根据不同payment有不同方法，比如paypal,就是固定BILLING,payssion 则对应pm_id",
			Order:      50},
		"order": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "付款按钮位置,1表示第一位",
			Order:      60},

		"payment_fee": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "支付收取的费用，有百分比，和固定费率两种，取决于fee_type",
			Order:      70},

		"fee_type": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "支付类型，1表示按照百分比收取，2表示固定收取费用，3表示超过某金额（helper)收取固定费用,选用3必须填入helper才能起作用",
			Order:      80},
		"helper": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "在第三种类型时，代表金额，超过这个金额，收取指定费用",
			Order:      100},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   220,
	}
}
