package paypal

import (
	"bytes"
	"errors"

	"encoding/gob"

	"encoding/json"
	"fmt"

	"github.com/agreyfox/eshop/payment/data"

	"github.com/agreyfox/eshop/system/admin"

	"strconv"
	"strings"
	"time"
)

// valid user request for payssion
func validateRequest(req *data.UserSubmitOrderRequest) error {

	if req.Amount <= 0 {
		return fmt.Errorf("amount data error")
	}
	if req.Amount < 0.0001 {
		return fmt.Errorf("amount data too small")
	}
	if len(req.Email) == 0 {
		return fmt.Errorf("no payer id")
	}
	if len(req.RequestInfo) == 0 {
		return fmt.Errorf("no user request info")
	}
	if len(req.Currency) == 0 {
		return fmt.Errorf("no user currency info")
	}
	return nil
}

// create skril
func createOrder(r *data.UserSubmitOrderRequest) (*Order, error) {
	logger.Debug("User create the paypal payment")

	invoiceID := data.GetShortOrderID()
	order := OrderRequest{
		Payer:    r.Payment,
		Email:    r.Email,
		Method:   r.PaymentChannel,
		Locale:   r.Locale,
		Comments: r.RequestInfo,
	}
	order.ItemList = []PurchaseUnitRequest{}
	pu := PurchaseUnitRequest{
		Amount: &PurchaseUnitAmount{
			Currency: r.Currency,
			Value:    fmt.Sprint(r.Amount),
		},
		InvoiceID:   invoiceID,
		Description: r.ContactInfo,
	}
	order.ItemList = append(order.ItemList, pu)

	orderPayer := &CreateOrderPayer{
		EmailAddress: order.Email,
	}

	//paymentList
	intent := OrderIntentCapture
	lo := order.Locale
	if len(lo) == 0 {
		lo = "en-US"
	}
	context := &ApplicationContext{
		ReturnURL:   returnURL,
		CancelURL:   cancelURL,
		BrandName:   brand_name,
		UserAction:  "PAY_NOW",
		LandingPage: LandingPageTypeLogin,
		Locale:      lo,
	}
	if order.Method == LandingPageTypeBilling {
		context.LandingPage = LandingPageTypeBilling // support credit card payment.
	}
	logger.Debugf("Request struct:%v", r)
	logger.Debug("Ready to create a order,Data Save to tempdb!")

	orderrespond, err := payClient.CreateOrder(intent, order.ItemList, orderPayer, context)
	return orderrespond, err

}

// to save a record
func saveOrderRequest(order Order, request *OrderRequest, ip string) error {
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
	da, _ := json.Marshal(request)
	record.Info = string(da)
	record.Description = getPurchaseContentFromRequest(request)
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

	record.BuyerEmail = request.Email // should support email
	/* if ok {
		record.BuyerEmail = email // string(email[:])
	} */
	record.RequestTime = time.Now().Unix()
	record.IP = ip

	s := data.SavePaymentLog(&record)
	if s {
		logger.Debug("Save payment log done! order id:", order.ID)
	} else {
		logger.Error("Save payment log error!,order id :", order.ID)
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

// create order by notification data and history data
func CreateNewOrderInDB(notifyData *WebHookNotifiedEvent, cap CaptureResource) (int, data.Order, bool) {
	var status string
	if cap.Status == PaypalPending {
		status = data.OrderCreated
	}
	if cap.Status == PaypalCompleted {
		status = data.OrderCompleted
	}
	ID := cap.InvoiceID

	paymentid := getOrderIDFromUrl(cap.Links)
	capdetail, _ := json.Marshal(cap)
	detail := string(capdetail)

	databin, _ := json.Marshal(notifyData)
	order := data.Order{

		Status: status,
		//OrderRequest:  record.Request,
		OrderDetail:   detail,
		OrderID:       ID,
		PaymentID:     paymentid,
		TransactionID: cap.ID,
		PaymentVendor: "paypal",
		PaymentMethod: notifyData.ResourceType,
		PaymentNote:   brand_name,
		NotifyInfo:    string(databin[:]),
		Description:   detail, //notifyData.Summary + "," + ID, // just for flil up incase no item list
		Currency:      cap.Amount.Currency,
		Total:         cap.Amount.Value,
		Paid:          cap.SellerPayableBreakdown.PayPalFee.Value,
		Net:           cap.SellerPayableBreakdown.NetAmount.Value,
		AdminNote:     "",
		UpdateTime:    fmt.Sprint(time.Now().Format(time.RFC1123)),
		Paytime:       fmt.Sprint(time.Now().Format(time.RFC1123)),
		IsRefund:      false,
		IsChargeBack:  false,
	}

	if cap.Status == PaypalCompleted {
		logger.Warn("Paypal Notified message is completed. checking exsiting order with order_id")
		// validreturn 0, data.Order{}, false
		oid, err := GetOrderByID(ID)
		id, err := strconv.ParseInt(oid, 10, 0)
		if err == nil {
			logger.Info("Found order with id:", id)
			ok := UpdateOrderStatusByID(oid, status)
			if ok {
				logger.Info("Update the existing order status to ", status)

			} else {
				logger.Error("Update the existing order status error,order_id is ", oid)

			}
			return int(id), order, ok
		}
	}
	//logger.Debug(time.Now())
	/* 	originReq, err := data.GetRequestByState(ID, PaypalCreated)
	   	//logger.Debug(time.Now())
	   	if err == nil {
	   		order.Payer = originReq.BuyerEmail
	   		order.PayerIP = originReq.IP
	   		order.Comments = originReq.Comments
	   		order.RequestTime = fmt.Sprint(time.Unix(originReq.RequestTime, 0).Format(time.RFC1123))
	   		order.Description = originReq.Description
		   } */
	request, err := data.GetRequestByID(ID)

	if err == nil {
		order.Payer = request.Email //.BuyerEmail
		order.PayerIP = request.IPAddr
		order.Comments = request.RequestInfo
		order.RequestTime = fmt.Sprint(time.Unix(request.OrderDate, 0).Format(time.RFC1123))
		purchaselist, _ := json.MarshalIndent(request.ItemList, "", "  ")
		order.Description = string(purchaselist)
	}
	order.Status = data.OrderPaid //标识已付
	mm, _ := json.Marshal(order)

	retcode, ok := admin.CreateContent("Order", mm)

	if ok {
		logger.Debug("Order created!", retcode)
		return retcode, order, true
	} else {
		return 0, data.Order{}, false
	}

}

func GetOrderByID(id string) (string, error) {

	oid := admin.FindContentID("Order", id, "order_id")
	// update the record
	if len(oid) > 0 {
		return oid, nil
	}
	return "", fmt.Errorf("not found")
}

// update order with order_id = id
func UpdateOrderStatusByID(id, state string) bool {

	oid := admin.FindContentID("Order", id, "order_id")
	// update the record
	if data.IsValidStatus(state) && len(oid) > 0 {
		_, err := admin.UpdateContent("Order", fmt.Sprintf("%s", oid), "state", []byte(state))
		if err != nil {
			logger.Error("update status error:", err)
			return false
		}
		return true
	}
	logger.Error("Not valid status!")
	return false
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

// 获得interface 对象中的bytes
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
	request, err := data.GetRequestByID(id)
	if err != nil {
		logger.Error("can not find order request id  ")
		return ""
	}
	purchaselist, _ := json.MarshalIndent(request.ItemList, "", "  ")
	return string(purchaselist)
}

// 生成购买物品的简单内容
func getPurchaseContentFromOrder(order Order) string {
	ret := ""
	logger.Debug("Order infor:========================")
	data.PrettyPrint(order)
	for i, ordeEntry := range order.PurchaseUnits {

		ret = ret + fmt.Sprintf("%d", i+1) + ":"
		for _, item := range ordeEntry.Items {
			ret = ret + item.Name + "--" + item.Description + "<br>"
		}
	}
	if ret == "" {
		ret = "DIGIT GOODS"
	}
	return ret
}

// 生成购买物品的简单内容
func getPurchaseContentFromRequest(request *OrderRequest) string {
	ret := ""
	logger.Debug("Request infor:========================")
	//PrettyPrint(request)

	for _, orderEntry := range request.ItemList {

		for j, item := range orderEntry.Items {
			logger.Debug("found item name ", item.Name)
			if len(item.Name) > 0 {
				ret += fmt.Sprintf("%d", j+1) + ":" + item.Name
				if len(item.Description) > 0 {
					ret += ":" + item.Description + "<br>"
				} else {
					ret += "<br>"
				}
			}
		}

	}

	if ret == "" {
		ret = "DIGIT GOODS"
	}
	return ret
}

// 生成购买物品的简单内容
func GeneratePurchaseContentFromRequest(request *PurchaseUnitRequest) string {
	ret := ""
	logger.Debug("Request infor:========================")
	data.PrettyPrint(request)
	for j, item := range request.Items {
		logger.Debug("found item name ", item.Name)
		if len(item.Name) > 0 {
			ret += fmt.Sprintf("%d", j+1) + ":" + item.Name
			if len(item.Description) > 0 {
				ret += ":" + item.Description + "<br>"
			} else {
				ret += "<br>"
			}
		}
	}

	if ret == "" {
		ret = "DIGIT GOODS"
	}
	return ret
}

func GetTransationDetail(id string) (SearchTransactionDetails, bool) {
	logger.Debug("==================search for id :", id)
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

		return SearchTransactionDetails{}, false
	} else {
		logger.Debug(err)
		return SearchTransactionDetails{}, false
	}

}

func getOrderRequestFromUserSubmit(userSubmit []byte) (*OrderRequest, error) {
	content := OrderRequest{}
	json.Unmarshal(userSubmit, &content)
	order := OrderRequest{}
	order.Email = content.Email
	order.Payer = content.Payer
	order.Method = content.Method
	local := content.Locale
	if len(local) > 0 {
		order.Locale = local
	} else {
		order.Locale = "en-US"
	}
	comment := content.Comments
	if len(comment) > 0 {
		order.Comments = comment
	}
	order.ItemList = content.ItemList
	if len(order.ItemList) == 0 {
		return &OrderRequest{}, errors.New("No payment list")
	}
	/* type purchaselist struct {
		List []interface{} `json:"item_list"`
	}

	listview := purchaselist{}
	json.Unmarshal(userSubmit, &listview)
	//err := json.Unmarshal([]byte(content["item_list"]), list)
	if len(listview.List) == 0 {
		return &OrderRequest{}, errors.New("No item")
	}
	fmt.Println(listview.List) */
	return &order, nil
}
