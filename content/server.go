package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Server struct {
	item.Item

	Name      string  `json:"name"`
	ShortName string  `json:"sName,omitempty"`    //长名
	LongName  string  `json:"longName,omitempty"` //长名
	Game      string  `json:"game"`
	Online    bool    `json:"online"`
	Category  string  `json:"category,omitempty"`
	Tags      string  `json:"tags,omitempty"`
	Coins     string  `json:"coins,omitempty"` //服务器上所有在卖的coin
	Items     string  `json:"items,omitempty"` //服务器上的所有在卖的item
	UnitPrice float32 `json:"price"`           // 金币单价
	UnitName  string  `json:"unitName"`
	desc      string  `json:"description,omitempty`
}

// MarshalEditor writes a buffer of html to edit a Server within the CMS
// and implements editor.Editable
func (s *Server) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(s,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Server field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", s, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Game", s, map[string]string{
				"label":       "Game",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Online", s, map[string]string{
				"label":       "Online",
				"type":        "text",
				"placeholder": "Enter the Online here",
			}),
		},
		editor.Field{
			View: editor.InputRepeater("Category", s, map[string]string{
				"label":       "Category",
				"type":        "text",
				"placeholder": "Enter the Online here",
			}),
		},
		editor.Field{
			View: editor.InputRepeater("Tags", s, map[string]string{
				"label":       "Tags",
				"type":        "text",
				"placeholder": "Enter the Tags here",
			}),
		},
		editor.Field{
			View: editor.InputRepeater("Product", s, map[string]string{
				"label":       "Product",
				"type":        "text",
				"placeholder": "Enter the Tags here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Server editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Server"] = func() interface{} { return new(Server) }
}

// String defines how a Server is printed. Update it using more descriptive
// fields from the Server struct type
func (s *Server) String() string {
	return fmt.Sprintf("Server: %s", s.Name)
}

func (o *Server) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"name": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      1,
		},
		"sName": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Help:       "游戏的短名，显示在后台管理列表中",
			Order:      5,
		},
		"longName": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      20,
		},
		"game": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Game"},
			Required:   true,
			Order:      30,
		},
		"category": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents/search?type=Category&q=[[game]]"},
			Required:   false,
			Order:      40,
		},
		"tags": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{"array"},
			Help:       "游戏的标签，可以输入多个，用空格分开",
			Order:      50,
		},
		"coins": {
			Type:       "multiselect",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Product", "array"},
			Help:       "本服务器只销售的金币类产品，为多选",
			Order:      60,
		},
		"items": {
			Type:       "multiselect",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Product", "array"},
			Help:       "本服务器在销售的道具，多选",
			Order:      70,
		},
		"online": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{""},
			Required:   true,
			Order:      80,
		},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{""},
			Order:      90,
		},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   30,
	}
}
