package content

import (
	"fmt"

	"github.com/agreyfox/eshop/system/item"
)

type Email struct {
	item.Item

	Name      string `json:"name"`
	Subject   string `json:"subject"`
	EmailBody string `json:"emailbody"`
	CC        string `json:"cc,omitempty"`
	Enable    bool   `json:"enable"`
	Desc      bool   `json:"description,omitempty"`
}

// MarshalEditor writes a buffer of html to edit a News within the CMS
// and implements editor.Editable
func (n *Email) MarshalEditor() ([]byte, error) {

	return nil, fmt.Errorf("Failed to render News editor view: %s", err.Error())

}

func init() {
	item.Types["Email"] = func() interface{} { return new(Email) }
}

// String defines how a News is printed. Update it using more descriptive
// fields from the News struct type
func (n *Email) String() string {
	return fmt.Sprintf("Email: %s", n.UUID)
}

func (o *Email) ContentStruct() map[string]interface{} {
	dd := map[string]item.FieldDescription{
		"name": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      10},
		"subject": {
			Type:       "input",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      20},
		"emailbody": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      30},
		"cc": {
			Type:       "input",
			DataType:   "field",
			DataSource: []string{},
			Order:      40},
		"enable": {
			Type:       "bool",
			DataType:   "field",
			Required:   true,
			DataSource: []string{},
			Order:      50},
		"description": {
			Type:       "textarea",
			DataType:   "field",
			DataSource: []string{},
			Order:      80},
	}
	//retStr, _ := json.Marshal(dd)
	return map[string]interface{}{
		"data": dd,
		"no":   260,
	}
}
