package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Product struct {
	item.Item

	Name   string  `json:"name"`
	Stock  uint    `json:"stock"`
	Desc   string  `json:"desc"`
	Logo   string  `json:"logo"`
	Type   uint    `json:"type"`
	Game   string  `json:"game"`
	Online bool    `json:"online"`
	Price  float32 `json:"price"`
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
