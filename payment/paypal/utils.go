package paypal

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/system/admin"

	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
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

	fmt.Println(string(bodyBytes[:]))

	if err = json.Unmarshal(bodyBytes, &t); err != nil {
		logger.Debug("Data looks like bad, Maybe it is not json data")
		//		logger.Debugf("%+v", err)
		return nil
	}
	return t
}

// to save a record
func saveOrderRequest(order Order, request map[string]interface{}, ip string) error {
	//logger.Debug(order)
	record := data.PaymentLog{
		RequestData: request,
		ReturnData:  order,
	}
	oid := ""
	comments := ""
	if len(order.PurchaseUnits) > 0 {
		item := order.PurchaseUnits[0]
		oid = item.InvoiceID
		comments = order.PurchaseUnits[0].Description
	} else {
		oid = order.ID
	}
	record.PaymentMethod = "paypal"
	record.OrderID = oid
	record.PaymentID = order.Intent
	record.PaymentState = order.Status
	da, _ := json.Marshal(order.PurchaseUnits)
	record.Info = string(da)
	record.Description = GetPurchaseContentFromOrder(order)
	record.Comments = comments
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

	email, ok := request["email"].(string) // should support email
	if ok {
		record.BuyerEmail = email // string(email[:])
	}
	record.RequestTime = time.Now().Unix()
	record.IP = ip

	s := data.SavePaymentLog(&record)
	if s {
		logger.Debug("Save payment log done! orderid:", order.ID)
	} else {
		logger.Error("Save payment log error!,orderid :", order.ID)
	}
	return nil
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

func CreateNewOrderInDB(notifyData *WebHookNotifiedEvent, cap CaptureResource) (int, data.Order, bool) {
	if cap.Status != PaypalCompleted {
		logger.Warn("Paypal Notified message is not completed.")
		return 0, data.Order{}, false
	}

	ID := cap.InvoiceID

	//val := &bytes.Buffer{}
	//states, err := data.GetLog(ID)
	//logger.Info(states)
	paymentid := getOrderIDFromUrl(cap.Links)
	detail := PrettyPrint(cap)
	//logger.Debugf("Invoice id is %s, order id is %s,paymentID is %s ", ID, ID, paymentid)
	databin, _ := json.Marshal(notifyData)
	order := data.Order{

		Status: data.OrderPaid,
		//OrderRequest:  record.Request,
		OrderDetail:   detail,
		OrderID:       ID,
		PaymentID:     paymentid,
		TransactionID: cap.ID,
		PaymentVendor: "paypal",
		PaymentMethod: notifyData.ResourceType,
		PaymentNote:   brand_name,
		NotifyInfo:    string(databin[:]),
		Description:   notifyData.Summary + "," + ID, // just for flil up incase no item list
		Currency:      cap.Amount.Currency,
		Total:         cap.Amount.Value,
		Paid:          "", //cap.SellerPayableBreakdown.PayPalFee.Value,
		Net:           "", //cap.SellerPayableBreakdown.NetAmount.Value,
		AdminNote:     "",
		UpdateTime:    fmt.Sprint(time.Now().Format(time.RFC1123)),
		Paytime:       fmt.Sprint(time.Now().Format(time.RFC1123)),
		IsRefund:      false,
		IsChargeBack:  false,
	}
	//logger.Debug(time.Now())
	originReq, err := data.GetRequestByState(ID, PaypalCreated)
	//logger.Debug(time.Now())
	if err == nil {
		order.Payer = originReq.BuyerEmail
		order.PayerIP = originReq.IP
		order.Comments = originReq.Comments
		order.RequestTime = fmt.Sprint(time.Unix(originReq.RequestTime, 0).Format(time.RFC1123))
		order.Description = originReq.Description
	}
	mm, _ := json.Marshal(order)

	retcode, ok := admin.CreateContent("Order", mm)

	//logger.Debugf("Find orderid by payment id is %s", admin.FindContentID("Order", cap.ID, "payment_id"))
	if ok {
		logger.Debug("Order created!", retcode)
		return retcode, order, true
	} else {
		return 0, data.Order{}, false
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

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetPurchaseContent via invoice id
func GetPurchaseContent(id string) string {
	ret := "Digital GOODS"
	record, err := data.GetLogContent(id, PaypalCreated)
	if err == nil {
		paymentlog := data.PaymentLog{}

		err := json.Unmarshal(record, &paymentlog)
		if err == nil {
			ret = paymentlog.Description
		}
		/*if err == nil {
			req := []byte(PrettyPrint(paymentlog.RequestData))
			sd := OrderRequest{}
			err = json.Unmarshal(req, &sd)
			if err == nil && len(sd.ItemList) > 0 {
				if len(sd.ItemList[0].Items) != 0 {
					ret = ""
					for _, item := range sd.ItemList[0].Items {
						ret = ret + item.Name + ":" + item.Description + "<br>"
					}
				}
			}
		} */
	}
	//logger.Debug(ret)
	return ret
}

// 生成购买物品的简单内容
func GetPurchaseContentFromOrder(order Order) string {
	ret := ""
	for i, ordeEntry := range order.PurchaseUnits {
		ret = ret + fmt.Sprint("%d", i+1) + ":"
		for _, item := range ordeEntry.Items {
			ret = ret + item.Name + "--" + item.Description + "<br>"
		}
	}
	if ret == "" {
		ret = "DIGIT GOODS"
	}
	return ret
}

func GetTransationDetail(id string) (SearchTransactionDetails, bool) {
	logger.Debug("=========================================================================search for id :", id)
	now := time.Now()
	last := now.AddDate(0, 0, -30) //两年内的单子
	//fmt.Print(now, last)
	page := 1
	pageSize := 5
	allField := "all"
	req := TransactionSearchRequest{
		//	TransactionID: &tid,
		EndDate:   now,
		StartDate: last,
		Page:      &page,
		PageSize:  &pageSize,
		Fields:    &allField,
	}
	if id == "" || id == "nil" {
		logger.Debug("Search Transaction without transactions id ")
	} else {
		req.TransactionID = &id
	}

	result, err := payClient.ListTransactions(&req)
	if err == nil {
		logger.Debug(result)
		if len(result.TransactionDetails) > 0 {
			return result.TransactionDetails[0], true
		}
		//PrettyPrint(result)
		return SearchTransactionDetails{}, false
	} else {
		logger.Debug(err)
		return SearchTransactionDetails{}, false
	}

}
