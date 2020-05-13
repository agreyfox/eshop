package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Category struct {
	item.Item

	Name     string `json:"name"`
	Game     string `json:"game"`
	Online   bool   `json:"online"`
	Belongto string `json:"belongto"`
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