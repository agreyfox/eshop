package content

import (
	"fmt"
	"net/http"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/system/item"
)

type Order struct {
	item.Item

	Email     string                   `json:"email"`
	Product   []map[string]interface{} `json:"product"`
	Bad       bool                     `json:"bad"`
	Status    uint                     `json:"status"`
	Total     float32                  `json:"total"`
	Due       uint64                   `json:"due"`
	Ispay     bool                     `json:"ispay"`
	Paytime   uint64                   `json:"paytime"`
	Payerinfo string                   `json:"payerinfo"`
	Worker    string                   `json:"worker"`
	Delivery  uint64                   `json:"delivery"`
	Social    string                   `json:"social"`
	Coupon    string                   `json:"coupon"`
}

// MarshalEditor writes a buffer of html to edit a Order within the CMS
// and implements editor.Editable
func (o *Order) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(o,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Order field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Email", o, map[string]string{
				"label":       "Email",
				"type":        "text",
				"placeholder": "Enter the Email here",
			}),
		},

		editor.Field{
			View: editor.Input("Bad", o, map[string]string{
				"label":       "Bad",
				"type":        "text",
				"placeholder": "Enter the Bad here",
			}),
		},
		editor.Field{
			View: editor.Input("Status", o, map[string]string{
				"label":       "Status",
				"type":        "text",
				"placeholder": "Enter the Status here",
			}),
		},
		editor.Field{
			View: editor.Input("Total", o, map[string]string{
				"label":       "Total",
				"type":        "text",
				"placeholder": "Enter the Total here",
			}),
		},
		editor.Field{
			View: editor.Input("Due", o, map[string]string{
				"label":       "Due",
				"type":        "text",
				"placeholder": "Enter the Due here",
			}),
		},
		editor.Field{
			View: editor.Input("Ispay", o, map[string]string{
				"label":       "Ispay",
				"type":        "text",
				"placeholder": "Enter the Ispay here",
			}),
		},
		editor.Field{
			View: editor.Input("Paytime", o, map[string]string{
				"label":       "Paytime",
				"type":        "text",
				"placeholder": "Enter the Paytime here",
			}),
		},
		editor.Field{
			View: editor.Input("Payerinfo", o, map[string]string{
				"label":       "Payerinfo",
				"type":        "text",
				"placeholder": "Enter the Payerinfo here",
			}),
		},
		editor.Field{
			View: editor.Input("Worker", o, map[string]string{
				"label":       "Worker",
				"type":        "text",
				"placeholder": "Enter the Worker here",
			}),
		},
		editor.Field{
			View: editor.Input("Delivery", o, map[string]string{
				"label":       "Delivery",
				"type":        "text",
				"placeholder": "Enter the Delivery here",
			}),
		},
		editor.Field{
			View: editor.Input("Social", o, map[string]string{
				"label":       "Social",
				"type":        "text",
				"placeholder": "Enter the Social here",
			}),
		},
		editor.Field{
			View: editor.Input("Coupon", o, map[string]string{
				"label":       "Coupon",
				"type":        "text",
				"placeholder": "Enter the Coupon here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Order editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Order"] = func() interface{} { return new(Order) }
}

// String defines how a Order is printed. Update it using more descriptive
// fields from the Order struct type
func (o *Order) String() string {
	return fmt.Sprintf("Order: %s", o.Email+"|"+o.Social)
}

func (o *Order) EnableSubContent() ([]string, bool) {
	return []string{"product"}, true
}

// EnableOwnerCheck, Indicate only get belone to user's content
func (o *Order) EnableOwnerCheck() bool {
	return true
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) error {
	logger.Debug("User create Order")

	return nil
}

//need get the product subcontent insert to retdata
func (o *Order) BeforeAPIResponse(w http.ResponseWriter, r *http.Request, retdata []byte) ([]byte, error) {
	logger.Debug("User retrieve Order")
	return retdata, nil
}

func (o *Order) Approve(w http.ResponseWriter, r *http.Request) error {
	logger.Debug("approve the order from pending to public")
	return nil
}

// enable autopprove
func (o *Order) AutoApprove(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (o *Order) AfterSave(w http.ResponseWriter, r *http.Request) error {
	logger.Debug("Process the sub buckets ")
	/* subData := map[string]interface{}{}
	obj := r.Header.Get("lqcms_json")
	id := r.Header.Get("lqcms_id")
	if obj != "" {
		err = json.Unmarshal([]byte(obj), &subData)
		if err != nil {
			logger.Error(err)
			return err
		}
		child := subData["product"].([]interface{})
		data := []map[string]interface{}{}
		for _, item := range child {
			one := item.(map[string]interface{})
			data = append(data, one)
		}
		if id == "" {
			logger.Error(err)
			return err
		}
		db.SetSubContent("Order"+api.PENDINGSuffix+":"+id, "product", data)
		gdd, err := db.GetSubContent("Order"+api.PENDINGSuffix+":"+id, "product")
		fmt.Println(err)
		if err == nil {
			mm := []map[string]interface{}{}
			errrrr := json.Unmarshal(gdd, &mm)

			fmt.Printf("%v,%v", mm, errrrr)
		}
	}
	*/
	return nil
}

func (o *Order) IndexContent() bool {
	return true
}
