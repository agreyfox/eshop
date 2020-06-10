package paypal

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/system/admin"

	"github.com/gofrs/uuid"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func getBinaryDataFromBody(req *http.Request) []byte {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(req.Body)
	}
	// Restore the io.ReadCloser to its original state
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes
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
		logger.Debug("Data looks like bad, Maybe it is not json data")
		logger.Debugf("%+v", err)
		return nil
	}
	return t
}

// to save a record
func saveOrderRequest(order Order, request map[string]interface{}, ip string) error {
	logger.Debug(order)
	record := data.PaymentLog{
		RequestData: request,
		ReturnData:  order,
	}
	oid := ""
	if len(order.PurchaseUnits) > 0 {
		item := order.PurchaseUnits[0]
		oid = item.InvoiceID
	} else {
		oid = order.ID
	}
	record.PaymentMethod = "paypal"
	record.OrderID = oid
	record.PaymentID = order.Intent
	record.PaymentState = order.Status
	da, _ := json.Marshal(order.PurchaseUnits)
	record.Info = string(da)

	var tt float64
	currency := ""
	for _, item := range order.PurchaseUnits {
		v2, err := strconv.ParseFloat(item.Amount.Value, 64)
		if err != nil {
			logger.Error(err)
			continue
		}
		tt = tt + v2
		currency = item.Amount.Currency
	}
	record.Total = fmt.Sprintf("%.2f", tt)
	record.Currency = currency

	email, ok := json.Marshal(request["payee"])
	if ok == nil {
		record.BuyerEmail = string(email[:])
	}
	record.RequestTime = time.Now().Unix()
	record.IP = ip

	s := data.SavePaymentLog(&record)
	if s {
		logger.Info("Save payment log done!")
	} else {
		logger.Info("Save payment log error!")
	}
	return ok
}

// to save a capture
func saveCapture(order CaptureOrderResponse) bool {
	record := data.PaymentLog{

		ReturnData: order,
	}
	record.PaymentMethod = "paypal"
	record.OrderID = order.ID
	//record.PaymentID = order.Intent
	record.PaymentState = order.Status
	da, _ := json.Marshal(order.PurchaseUnits)
	record.Info = string(da)

	currency := "-"

	record.Total = currency
	record.Currency = currency

	record.BuyerEmail = order.Payer.EmailAddress

	record.RequestTime = time.Now().Unix()

	s := data.SavePaymentLog(&record)
	if s {
		logger.Info("Save payment log done!")
	} else {
		logger.Info("Save payment log error!")
	}
	return s
}

func getOrderID() string {
	uid, err := uuid.NewV1()
	if err != nil {
		return ""
	}
	hash := md5.Sum([]byte(uid.String()))

	return hex.EncodeToString(hash[:])
	//return uid.String()
}

// to save a record
func saveNotify(t WebHookNotifiedEvent) bool {
	// Store the user model in the user bucket using the username as the key.

	record := data.PaymentLog{
		ReturnData: t,
	}

	record.OrderID = t.ID
	record.PaymentMethod = "paypal"
	record.PaymentID = t.EventType
	record.Total = ""
	record.Currency = ""
	record.PaymentState = "Notify-" + t.EventType

	record.RequestTime = t.CreateTime.Unix()
	record.Info = t.Summary
	ok := data.SavePaymentLog(&record)
	if ok {
		logger.Info("Save Notify log done!")
	} else {
		logger.Info("Save Notify log error!")
	}
	return ok

}

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
func PrettyPrint(obj interface{}) string {

	prettyJSON, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		logger.Fatal("Failed to generate json", err)
	}
	fmt.Println("===================================================================")
	fmt.Printf("%s\n", string(prettyJSON))
	fmt.Println("===================================================================")
	return string(prettyJSON)
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

// get orderid from link
func getOrderIDFromUrl(link []Link) string {
	if len(link) >= 1 {
		for _, link := range link {
			if link.Rel == "up" {
				pp := strings.Split(link.Href, "/")
				if len(pp) > 5 {
					return pp[len(pp)-1]
				} else {
					return ""
				}
			}
		}
	}
	return ""
}

func CreateNewOrderInDB(notifyData *WebHookNotifiedEvent, cap CaptureResource) (int, bool) {
	if cap.Status != PaypalCompleted {
		logger.Warn("Paypal Notified message is not completed.")
		return 0, false
	}

	ID := cap.InvoiceID

	//val := &bytes.Buffer{}
	//states, err := data.GetLog(ID)
	//logger.Info(states)
	oid := getOrderIDFromUrl(cap.Links)
	detail := PrettyPrint(cap)
	logger.Debugf("Invoice id is %s, order id is %s,paymentID is %s ", ID, oid, cap.ID)
	databin, _ := json.Marshal(notifyData)
	order := data.Order{

		Status: data.OrderPaid,
		//OrderRequest:  record.Request,
		OrderDetail:   detail,
		OrderID:       oid,
		PaymentID:     cap.ID,
		PaymentVendor: "paypal",
		PaymentMethod: notifyData.ResourceType,
		PaymentNote:   brand_name,
		NotifyInfo:    string(databin[:]),
		Description:   notifyData.Summary,
		Comments:      "",
		Currency:      cap.Amount.Currency,
		Total:         cap.Amount.Value,
		Paid:          "", //cap.SellerPayableBreakdown.PayPalFee.Value,
		Net:           "", //cap.SellerPayableBreakdown.NetAmount.Value,
		AdminNote:     ID,
		UpdateTime:    fmt.Sprint(time.Now().Format(time.RFC1123)),
		Paytime:       fmt.Sprint(time.Now().Format(time.RFC1123)),
		IsRefund:      false,
		IsChargeBack:  false,
	}
	//logger.Debug(time.Now())
	originReq, err := data.GetRequestByState(oid, PaypalCreated)
	//logger.Debug(time.Now())
	if err == nil {
		order.Payer = originReq.BuyerEmail
		order.PayerIP = originReq.IP
		order.RequestTime = fmt.Sprint(time.Unix(originReq.RequestTime, 0).Format(time.RFC1123))
	}
	mm, _ := json.Marshal(order)

	retcode, ok := admin.CreateContent("Order", mm)
	logger.Debugf("Find orderid by payment id is %s", admin.FindContentID("Order", cap.ID, "payment_id"))
	if ok {
		return retcode, true
	} else {
		return 0, false
	}

}

// update order with payment_id = id
func UpdateOrderStatusByPaymentID(id, state string) bool {
	// find paymenid is id record

	oid := admin.FindContentID("Order", id, "payment_id")
	// update the record
	if data.IsValidStatus(state) && len(oid) > 0 {
		_, err := admin.UpdateContent("Order", fmt.Sprintf("%s", oid), "state", []byte(state))
		if err != nil {
			logger.Debug("update status error:", err)
			return false
		}
		return true
	}
	logger.Error("Not valid status!")
	return false
}

// update order with payment_id = id
func UpdateOrderStatusByOrderID(id, state string) bool {
	// find paymenid is id record

	oid := admin.FindContentID("Order", id, "order_id")
	// update the record
	if data.IsValidStatus(state) && len(oid) > 0 {
		_, err := admin.UpdateContent("Order", fmt.Sprintf("%s", oid), "state", []byte(state))
		if err != nil {
			logger.Debug("update status error:", err)
			return false
		}
		return true
	}
	logger.Error("Not valid status!")
	return false
}
