package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Product struct {
	item.Item
	Name            string  `json:"name"`
	Sname           string  `json:"sname"`
	Hot             bool    `json:"hotItem,omitempty"`     //是否在hotitem 中显示
	Stock           uint    `json:"stock"`                 //库存数量
	Desc            string  `json:"description,omitempty"` //html
	Logo            string  `json:"logo,omitempty"`        //图标文件
	Type            string  `json:"type"`                  //coin,item 两种
	Game            string  `json:"game"`
	Online          bool    `json:"online"`
	Price           float32 `json:"price"`                     //单价
	MN              uint    `json:"miniNumber"`                //最小购买数
	Unit            string  `json:"Unit"`                      //购买数量的单位
	HintImage       string  `json:"hintImage,omitempty"`       //提示图片
	HintText        string  `json:"hintText,omitempty"`        //提示文字
	PurchaseLabel   string  `json:"customerLabel"`             //用户输入提示内容
	PurchaseCaution string  `json:"customerCaution,omitempty"` //用户输入要求购买内容
	Discount        string  `json:"discount,omitempty"`        //使用discount模板

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
		"sname": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "内部使用的名字说明",
			Order:      5,
		},
		"hotItem": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{},
			Help:       "本产品是否出现在hotitem 栏目，选择是表示显示",
			Order:      20,
			Others:     "false",
		},
		"stock": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      30,
			Help:       "输入库存数量",
			Others:     "99999",
		},
		"game": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Game"},
			Required:   true,
			Order:      40,
		},
		"type": {
			Type:       "select",
			DataType:   "field",
			DataSource: []string{"coin", "item"},
			Required:   true,
			Help:       "产品类型，金币类还是道具类",
			Order:      50,
		},
		"price": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "产品的单价",
			Required:   true,
			Order:      52,
		},
		"discount": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Discount"},
			Required:   false,
			Help:       "若是金币类别，可使用系统中定义的discount为客户提供折扣，若为\n道具类，则这个字段无意义",
			Order:      70,
		},
		"online": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      80,
		},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      190,
		},
		"logo": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Help:       "产品的小方图片",
			Order:      150,
		},
		"hintImage": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Help:       "用户鼠标移动到本产品图片上方时，显示改产品的详细说明，也是一张图片",
			Order:      110,
		},
		"hintText": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "当鼠标移动到产品图片时，也可以显示文字说明，该文字说明和将显示在hintImage上方",
			Order:      120,
		},
		"miniNumber": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Help:       "本产品销售时，最小单位的数量",
			Order:      90,
		},
		"Unit": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Help:       "本产品销售时，最小单位的数量的单位，用于网页显示",
			Order:      56,
		},
		"customerLabel": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Help:       "当用户购买本产品时，提示用户输入购买要求，在输入框的下方",
			Order:      130,
		},
		"customerCaution": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "当用购买本产品时，提示用户购买注意事项，在输入框下方",
			Order:      140,
		},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   40,
	}
}
