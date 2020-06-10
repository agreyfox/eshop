package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Game struct {
	item.Item

	Name   string `json:"name"`
	Sname  string `json:"sname"`
	Logo   string `json:"logo,omitempty"`
	Image  string `json:"barImage,omitempty"`
	Online bool   `json:"online"`
	Desc   string `json:"description,omitempty"`
	Coupon string `json:"coupon,omitempty"` //使用那个coupon
	Hot    bool   `json:"hot"`              //是否在hotgame中显示
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
			Order:      1},
		"sname": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      2},
		"hot": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      3},
		"logo": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      6},
		"barImage": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      5},
		"coupon": {
			Type:       "select",
			DataType:   "field",
			DataSource: []string{"array"},
			Order:      4},
		"online": {
			Type:       "bool",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      4},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      7},
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
