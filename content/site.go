package content

import (
	"fmt"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Site struct {
	item.Item

	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Desc        string `json:"desc"`
}

// MarshalEditor writes a buffer of html to edit a Site within the CMS
// and implements editor.Editable
func (s *Site) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(s,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Site field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", s, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Owner", s, map[string]string{
				"label":       "Owner",
				"type":        "text",
				"placeholder": "Enter the Owner here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Site editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Site"] = func() interface{} { return new(Site) }
	//fmt.Println("Initial content site ")
}

// String defines how a Site is printed. Update it using more descriptive
// fields from the Site struct type
func (s *Site) String() string {
	return fmt.Sprintf("Site: %s", s.UUID)
}

func (s *Site) IndexContent() bool {
	return true
}

func (o *Site) ContentStruct() map[string]interface{} {
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
			Help:       "direct link显示出来的名称",
			Order:      5,
		},
		"order": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Required:   false,
			Order:      40,
		},
		"desc": {
			Type:       "textarea",
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
