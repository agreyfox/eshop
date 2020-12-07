package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type DirectGame struct {
	item.Item

	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Game        string `json:"game"`
	ItemType    string `json:"type,omitempty"`
	Order       int    `json:"order,omitempty"`
}

// MarshalEditor writes a buffer of html to edit a DirectGame within the CMS
// and implements editor.Editable
func (d *DirectGame) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(d,
		// Take note that the first argument to these Input-like functions
		// is the string version of each DirectGame field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", d, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("DisplayName", d, map[string]string{
				"label":       "DisplayName",
				"type":        "text",
				"placeholder": "Enter the DisplayName here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render DirectGame editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["DirectGame"] = func() interface{} { return new(DirectGame) }
}

// String defines how a DirectGame is printed. Update it using more descriptive
// fields from the DirectGame struct type
func (d *DirectGame) String() string {
	return fmt.Sprintf("DirectGame: %s", d.UUID)
}

func (o *DirectGame) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"name": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Order:      1,
		},
		"display_name": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   true,
			Help:       "direct显示出来的名称",
			Order:      5,
		},
		"game": {
			Type:       "select",
			DataType:   "content",
			DataSource: []string{"/admin/v1/contents?type=Game&count=-1"},
			Required:   true,
			Order:      30,
		},
		"order": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   false,
			Order:      40,
		},
		"type": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{""},
			Order:      90,
		},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   230,
	}
}
