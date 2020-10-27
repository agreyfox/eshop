package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Category struct {
	item.Item

	Name     string `json:"name"`
	Sname    string `json:"sname"`
	Game     string `json:"game"`
	Online   bool   `json:"online"`
	Belongto string `json:"belongto,omitempty"`
	Desc     string `json:"description,omitempty"`
	Icon     string `json:"icon,omitempty"`
	Class    string `json:"class"` //是金币还是物品
	Hint     string `json:"hint,omitempty"`
}

// MarshalEditor writes a buffer of html to edit a Category within the CMS
// and implements editor.Editable
func (c *Category) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(c,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Category field, and must follow
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
				"label": "Game",
				"type":  "text",
			}),
		},
		editor.Field{
			View: editor.Input("Online", c, map[string]string{
				"label":       "Online",
				"type":        "text",
				"placeholder": "Enter the Online here",
			}),
		},
		editor.Field{
			View: editor.Input("Belongto", c, map[string]string{
				"label":       "Belongto",
				"type":        "text",
				"placeholder": "Enter the father category here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Category editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Category"] = func() interface{} { return new(Category) }
}

// String defines how a Category is printed. Update it using more descriptive
// fields from the Category struct type
func (c *Category) String() string {
	return fmt.Sprintf("Category: %s", c.UUID)
}

func (o *Category) ContentStruct() map[string]interface{} {
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
			Help:       "名字描述，用于内部",
			Required:   true,
			Order:      5,
		},
		"game": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Game&count=-1"},
			Required:   true,
			Order:      20},
		"class": {
			Type:       "select",
			DataType:   "field",
			DataSource: []string{"coin", "item"},
			Required:   true,
			Help:       "目录类型，选择icon,或者item",
			Order:      22},
		"online": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      30},
		"belongto": {
			Type:       "tree",
			DataType:   "content",
			DataSource: []string{"tree"},
			Order:      40,
		},
		"hint": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      50},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      60},
		"icon": {
			Type:       "file",
			DataType:   "field",
			DataSource: []string{},
			Order:      70},
	}
	return map[string]interface{}{
		"data": dd,
		"no":   20,
	}

}
