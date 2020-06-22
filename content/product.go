package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Product struct {
	item.Item
	Name            string  `json:"name"`
	Hot             bool    `json:"hotItem"`     //是否在hotitem 中显示
	Stock           uint    `json:"stock"`       //库存数量
	Desc            string  `json:"description"` //html
	Logo            string  `json:"logo"`        //图标文件
	Type            uint    `json:"type"`        //coin,item 两种
	Game            string  `json:"game"`
	Online          bool    `json:"online"`
	Price           float32 `json:"price"`               //单价
	HintImage       string  `json:"hintImage,omitempty"` //提示图片
	HintText        string  `json:"hintText,omitempty"`  //提示文字
	MN              int     `json:"miniNumber"`          //最小购买数
	PurchaseLabel   string  `json:"customerLabel"`       //用户输入提示内容
	PurchaseCaution string  `json:"customerCaution"`     //用户输入要求购买内容
	Discount        string  `json:"discount,omitempty"`  //使用discount模板
}

// MarshalEditor writes a buffer of html to edit a Product within the CMS
// and implements editor.Editable
func (p *Product) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(p,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Product field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", p, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Stock", p, map[string]string{
				"label":       "Stock",
				"type":        "text",
				"placeholder": "Enter the Stock here",
			}),
		},
		editor.Field{
			View: editor.Textarea("Desc", p, map[string]string{
				"label":       "Desc",
				"placeholder": "Enter the Desc here",
			}),
		},
		editor.Field{
			View: editor.File("Logo", p, map[string]string{
				"label":       "Logo",
				"placeholder": "Upload the Logo here",
			}),
		},
		editor.Field{
			View: editor.Input("Type", p, map[string]string{
				"label":       "Type",
				"type":        "text",
				"placeholder": "Enter the Type here",
			}),
		},
		editor.Field{
			View: editor.Input("Game", p, map[string]string{
				"label":       "Game",
				"type":        "text",
				"placeholder": "Enter the game name here",
			}),
		},
		editor.Field{
			View: editor.Input("Online", p, map[string]string{
				"label":       "Online",
				"type":        "text",
				"placeholder": "Enter the Online here",
			}),
		},
		editor.Field{
			View: editor.Input("Price", p, map[string]string{
				"label":       "Price",
				"type":        "text",
				"placeholder": "Enter the Price here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Product editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Product"] = func() interface{} { return new(Product) }
}

// String defines how a Product is printed. Update it using more descriptive
// fields from the Product struct type
func (p *Product) String() string {
	return fmt.Sprintf("Product: %s", p.Name)
}

func (p *Product) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{

		"name": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      1,
		},
		"hotItem": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{},
			Order:      2,
			Others:     "false",
		},
		"stock": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      3,
			Others:     "99999",
		},
		"game": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Game"},
			Required:   true,
			Order:      4,
		},
		"type": {
			Type:       "select",
			DataType:   "field",
			DataSource: []string{"coin", "item"},
			Required:   true,
			Order:      4,
		},
		"price": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      5,
		},
		"discount": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Discount"},
			Required:   false,
			Order:      7,
		},
		"online": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      8,
		},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      19,
		},
		"logo": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      9,
		},
		"hintImage": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      10,
		},
		"hintText": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      11,
		},
		"miniNumber": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      6,
		},
		"customerLabel": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      12,
		},
		"customerCaution": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      13,
		},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   40,
	}
}
