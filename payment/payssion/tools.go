package payssion

import (
	"encoding/json"
	"fmt"

	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/payment/utils"
	"github.com/agreyfox/eshop/system/admin"

	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func getJSONFromBody(req *http.Request) map[string]interface{} {
	var body string
	bodyBytes, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		body = fmt.Sprintf("failed read Read response body: %v", err)
		logger.Debug(body)
		return nil
	}
	var t map[string]interface{}

	//fmt.Println(body)

	if err = json.Unmarshal(bodyBytes, &t); err != nil {
		logger.Debug("json data looks like bad")
		logger.Debugf("%+v", err)
		return nil
	}
	return t
}

// to save a record
func saveRequest(t *PaymentResponse, request map[string]interface{}) bool {
	// Store the user model in the user bucket using the username as the key.

	record := data.PaymentLog{
		RequestData: request,
		ReturnData:  t,
	}
	record.OrderID = t.Transaction.OrderID
	record.PaymentMethod = "payssion"
	record.PaymentID = t.Transaction.PMID
	record.Total = t.Transaction.Amount
	record.Currency = t.Transaction.Currency
	record.PaymentState = t.Transaction.State
	email, ok := request["payer_email"].(string)
	if ok {
		record.BuyerEmail = email
	}
	record.RequestTime = time.Now().Unix()

	buyComments, ok := request["description"].(string)
	if ok {
		record.Info = buyComments
	}
	ip, ok := request["ip"].(string)
	if ok {
		record.IP = ip
	}
	ok = data.SavePaymentLog(&record)
	if ok {
		logger.Info("Save payment log done!")
	} else {
		logger.Info("Save payment log error!")
	}
	return ok

}

// to save a record
func saveNotify(t *NotifyResponse) bool {
	// Store the user model in the user bucket using the username as the key.

	record := data.PaymentLog{
		ReturnData: t,
	}
	record.OrderID = t.OrderID
	record.PaymentMethod = "payssion"
	record.PaymentID = t.PMID
	record.Total = t.Amount
	record.Currency = t.Currency
	record.PaymentState = t.State

	record.RequestTime = time.Now().Unix()
	record.Info = t.Description
	ok := data.SavePaymentLog(&record)
	if ok {
		logger.Info("Save payment log done!")
	} else {
		logger.Info("Save payment log error!")
	}
	return ok

}

/*
// to save a capture
func saveCapture(order CaptureOrderResponse) error {
	// Store the user model in the user bucket using the username as the key.
	err := tempDBHandler.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(dbName))

		if err != nil {
			return err
		}

		encoded, err := json.Marshal(order)
		if err != nil {
			return err
		}
		id := order.Transaction.ID + "-" + order.Transaction.OrderID + "-capture"
		return b.Put([]byte(id), encoded)
	})
	return err
}
*/
//将interface 简单传回
func renderJSON(w http.ResponseWriter, r *http.Request, data interface{}) (int, error) {

	marsh, err := json.Marshal(data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

// print pretty map[string]internface output
func PrettyPrint(obj interface{}) {

	prettyJSON, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		logger.Fatal("Failed to generate json", err)
	}
	fmt.Println("===================================================================")
	fmt.Printf("%s\n", string(prettyJSON))
	fmt.Println("===================================================================")
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func getOrderID() string {
	return utils.RandomString(10)
	/* uid, err := uuid.NewV1()
	if err != nil {
		return ""
	}
	hash := md5.Sum([]byte(uid.String()))

	return hex.EncodeToString(hash[:]) */
	//return uid.String()
}

// CreateNewOrderInDB try to create new order in backend system with the information provide data and the record
/*
 system db structure is
	orderid:order structure
*/
func CreateNewOrderInDB(notifyData *NotifyResponse) (int, data.Order, bool) {

	oid := notifyData.OrderID
	ID := oid
	//val := &bytes.Buffer{}
	states, err := data.GetLog(ID)
	logger.Info(states)
	buff, _ := json.Marshal(notifyData)
	order := data.Order{

		Status: data.OrderPaid,
		//OrderRequest:  record.Request,
		//OrderDetail:   record.Result,
		OrderID:       oid,
		PaymentVendor: "payssion",
		PaymentMethod: notifyData.PMID,
		PaymentNote:   notifyData.AppName,
		NotifyInfo:    string(buff[:]),
		Description:   notifyData.Description,
		Comments:      notifyData.Description,
		Currency:      notifyData.Currency,
		Total:         notifyData.Amount,
		Paid:          notifyData.Paid,
		Net:           notifyData.Net,
		AdminNote:     oid,
		UpdateTime:    fmt.Sprint(time.Now().Format(time.RFC1123)),
		Paytime:       fmt.Sprint(time.Now().Format(time.RFC1123)),
		IsRefund:      false,
		IsChargeBack:  false,
	}

	originReq, err := data.GetRequestByState(ID, PayssionPending)

	if err == nil {
		databuff, _ := json.Marshal(originReq.ReturnData)
		order.OrderDetail = string(databuff[:])
		order.Payer = originReq.BuyerEmail
		order.PayerIP = originReq.IP
		order.RequestTime = fmt.Sprint(time.Unix(originReq.RequestTime, 0).Format(time.RFC1123))
	}

	mm, _ := json.Marshal(order)

	retCode, ok := admin.CreateContent("Order", mm)
	/*
		errd := data.SystemDBHandler.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(data.OrderName))

			if err != nil {
				return err
			}

			encoded, err := json.Marshal(order)
			if err != nil {
				return err
			}

			return b.Put([]byte(ID), encoded)

		})

		if errd != nil {
			logger.Error("Error when save order infor", errd)
			return false
		} */
	//	logger.Infof("Create order with id %s in system", ID)
	admin.UpdateContent("Order", fmt.Sprintf("%d", retCode), "status", []byte(data.OrderInValidate))
	if ok {
		return retCode, order, true
	} else {
		logger.Error("error in creat order with code :", retCode)
		return 0, order, false
	}

}

func UpdateOrderStatus(tid, oid, state string) bool {

	eid := admin.FindContentID("Order", oid, "order_id")
	// update the record
	if data.IsValidStatus(state) && len(eid) > 0 {
		_, err := admin.UpdateContent("Order", fmt.Sprintf("%s", eid), "status", []byte(state))
		if err != nil {
			logger.Debug("update Order status error:", err)
			return false
		}
		return true
	}

	return false

}
