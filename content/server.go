package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Server struct {
	item.Item

	Name     string `json:"name"`
	LongName string `json:"longName,omitempty"` //长名
	Game     string `json:"game"`
	Online   bool   `json:"online"`
	Category string `json:"category,omitempty"`
	Tags     string `json:"tags,omitempty"`
	Coins    string `json:"coins,omitempty"` //服务器上所有在卖的coin
	Items    string `json:"items,omitempty"` //服务器上的所有在卖的item
	desc     string `json:"description,omitempty`
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
		"longName": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      2,
		},
		"game": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Game"},
			Required:   true,
			Order:      3,
		},
		"category": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Category", "array"},
			Required:   false,
			Order:      4,
		},
		"tags": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{"array"},
			Help:       "游戏的标签，可以输入多个，用空格分开",
			Order:      5,
		},
		"coins": {
			Type:       "multiselect",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Product", "array"},
			Help:       "本服务器只销售的金币类产品，为多选",
			Order:      6,
		},
		"items": {
			Type:       "multiselect",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Product", "array"},
			Help:       "本服务器在销售的道具，多选",
			Order:      7,
		},
		"online": {
			Type:       "bool",
			DataType:   "field",
			DataSource: []string{""},
			Required:   true,
			Order:      8,
		},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{""},
			Order:      9,
		},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   30,
	}
}
