package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Server struct {
	item.Item

	Name     string   `json:"name"`
	Game     string   `json:"game"`
	Online   bool     `json:"online"`
	Category []string `json:"category"`
	Tags     []string `json:"tags"`
	Product  []string `json:"product"`
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
