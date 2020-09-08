package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Game struct {
	item.Item

	Name        string `json:"name"`
	Sname       string `json:"sname,omitempty"`
	Logo        string `json:"logo,omitempty"`
	Image       string `json:"barImage,omitempty"`
	Online      bool   `json:"online"`
	Hotitem     bool   `json:"hotitem"`
	CoinCode    string `json:"coinName"`
	ItemCode    string `json:"itemName"`
	ProductSell string `json:"productSell"` //  设定该游戏销售那些内容，coin,item，both
	Desc        string `json:"description,omitempty"`
	Coupon      string `json:"coupon,omitempty"` //使用那个coupon
	Hot         bool   `json:"hot,omitempty"`    //是否在hotgame中显示
}

// MarshalEditor writes a buffer of html to edit a Game within the CMS
// and implements editor.Editable
func (g *Game) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(g,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Game field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", g, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Sname", g, map[string]string{
				"label":       "Sname",
				"type":        "text",
				"placeholder": "Enter the Sname here",
			}),
		},
		editor.Field{
			View: editor.File("Logo", g, map[string]string{
				"label":       "Logo",
				"placeholder": "Upload the Logo here",
			}),
		},
		editor.Field{
			View: editor.Input("Online", g, map[string]string{
				"label":       "Online",
				"type":        "text",
				"placeholder": "Enter the Online here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Game editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Game"] = func() interface{} { return new(Game) }
}

func (o *Game) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"name": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Help:       "显示在用户界面中的名称",
			Order:      1},
		"sname": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      2},
		"coinName": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "用户界面中的coin类型名称",
			Required:   true,
			Order:      10},
		"itemName": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Help:       "用户界面中的物品类型名称",
			Required:   true,
			Order:      11},
		"productSell": {
			Type:       "select",
			DataType:   "field",
			DataSource: []string{"coin", "item", "both"},
			Help:       "控制用户界面中的本游戏销售物品，取值coin,item,both",
			Required:   true,
			Order:      15},
		"hot": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Help:       "选择本游戏是否出现在hot game列表中",
			Order:      30},
		"logo": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Help:       "用于游戏列中Icon的显示，大小：",
			Order:      40},
		"barImage": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Help:       "用于游戏页中的横幅显示，大小：",
			Order:      50},
		"coupon": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Coupon"},
			Help:       "选择当前激活的coupon",
			Order:      60},
		"online": {
			Type:       "bool",
			DataType:   "field",
			Required:   true,
			Help:       "true为激活游戏",
			DataSource: []string{},
			Order:      70},
		"hotitem": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{},
			Help:       "选择本游戏是否出现在hotitem列表中",
			Order:      80,
		},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      90},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   10,
	}
}

// String defines how a Game is printed. Update it using more descriptive
// fields from the Game struct type
func (g *Game) String() string {
	return fmt.Sprintf("Game: %s", g.Name)
}
