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
	Logo   string `json:"logo"`
	Online bool   `json:"online"`
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

// String defines how a Game is printed. Update it using more descriptive
// fields from the Game struct type
func (g *Game) String() string {
	return fmt.Sprintf("Game: %s", g.Name)
}
