package paypal

import (
	"bytes"
	"errors"
	"net/url"

	"encoding/gob"

	"encoding/json"
	"fmt"

	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/system/admin"
	"github.com/agreyfox/eshop/system/db"
	"github.com/boltdb/bolt"

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
		Description: r.RequestInfo,
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

// to save a record   , obsleted !
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
	// Store the user model in the user bucket using the usernam89B6249286542084B
	record := data.PaymentLog{
		ReturnData: t,
	}

	record.OrderID = t.ID + "-" + t.EventType // this is log id format
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
func CreateNewOrderInDB(notifyData *WebHookNotifiedEvent, cap Resource) (int, data.Order, bool) {
	logger.Debug("Create new order process start")
	var status string
	if cap.Status == PaypalPending {
		status = data.OrderCreated
	}
	if cap.Status == PaypalCompleted {
		status = data.OrderCompleted
	}
	ID := cap.PurchaseUnits[0].InvoiceID // our order id
	logger.Debugf("notified result is %v", cap)

	paymentid := getOrderIDFromUrl(cap.Links)
	capdetail, _ := json.Marshal(cap.PurchaseUnits)
	detail := string(capdetail)

	databin, _ := json.Marshal(notifyData)
	TransactionId := cap.ID //PurchaseUnits[0].Payments.Captures[0].ID
	logger.Warnf("Transaction id is %s", TransactionId)
	order := data.Order{

		Status: status,
		//OrderRequest:  record.Request,
		OrderDetail:   detail,
		OrderID:       ID,
		PaymentID:     paymentid,
		TransactionID: TransactionId,
		PaymentVendor: "paypal",
		PaymentMethod: notifyData.ResourceType,
		PaymentNote:   brand_name,
		NotifyInfo:    string(databin[:]),
		Description:   detail, //notifyData.Summary + "," + ID, // just for flil up incase no item list
		Currency:      cap.PurchaseUnits[0].Amount.Currency,
		Total:         cap.PurchaseUnits[0].Amount.Value,
		//	Paid:          cap.PurchaseUnits[0].Payments.Captures[0].SellerPayableBreakdown.PayPalFee.Value,
		//Net:          cap.PurchaseUnits[0].Payments.Captures[0].SellerPayableBreakdown.NetAmount.Value,
		AdminNote:    "",
		UpdateTime:   fmt.Sprint(time.Now().Format(time.RFC1123)),
		Paytime:      fmt.Sprint(time.Now().Format(time.RFC1123)),
		IsRefund:     false,
		IsChargeBack: false,
		Payer:        cap.Payer.EmailAddress, // add 2020/11/25
	}
	if len(cap.PurchaseUnits[0].Payments.Captures) > 0 && cap.PurchaseUnits[0].Payments.Captures[0].SellerPayableBreakdown != nil {
		order.Paid = cap.PurchaseUnits[0].Payments.Captures[0].SellerPayableBreakdown.PayPalFee.Value
		order.Net = cap.PurchaseUnits[0].Payments.Captures[0].SellerPayableBreakdown.NetAmount.Value
	}
	request, err := data.GetRequestByID(ID)

	if err == nil {
		order.User = request.Email //.BuyerEmail
		order.PayerIP = request.IPAddr
		order.PayerLink = request.ContactInfo
		order.Comments = request.RequestInfo
		order.RequestTime = fmt.Sprint(time.Unix(request.OrderDate, 0).Format(time.RFC1123))
		purchaselist, _ := json.MarshalIndent(request.ItemList, "", "  ")
		order.Description = string(purchaselist)
	} else {
		logger.Errorf("The request is not found, This is wired!,ID is %f", ID)
	}

	if cap.Status == PaypalCompleted {
		logger.Warn("Paypal Notified message is completed. checking exsiting order with order_id:%s", ID)
		// validreturn 0, data.Order{}, false
		oid, err := GetIdByOrderID(ID)

		id, err := strconv.ParseInt(oid, 10, 0)
		if err == nil {
			logger.Infof("Found order with id %s need to update status:", oid)
			ok := UpdateOrderByOrderID(ID, &order)
			/* ok := UpdateOrderStatusByID(oid, status)
			aa, err := admin.UpdateContent("Order", oid, "transaction_id", []byte(order.TransactionID))
			aa, err = admin.UpdateContent("Order", oid, "payer_link", []byte(order.PayerLink))
			aa, err = admin.UpdateContent("Order", oid, "notify_info", []byte(order.NotifyInfo))
			//net
			//Paid
			//paytime */
			logger.Debug("Update order with muilti Field:", oid, order.TransactionID, order.PayerLink)
			if ok {
				logger.Info("Update the existing order status to ", status)

			} else {
				logger.Error("Update the existing order status error,order_id is ", oid)

			}
			return int(id), order, ok
		} else {
			logger.Warn("the previous order is not found , will continue to create new order.", ID)
		}
	}
	order.Status = data.OrderPaid //标识已付
	mm, _ := json.Marshal(order)

	retcode, ok := admin.CreateContent("Order", mm)

	if ok {
		logger.Debug("Order created!", retcode)
		return retcode, order, true
	} else {
		logger.Errorf("Order created error %s,%s!", ID, retcode)
		return 0, data.Order{}, false
	}

}

// create pending order by notification data and history data
func CreateNewPendingOrderInDB(notifyData *WebHookNotifiedEvent, cap Resource) (int, data.Order, bool) {
	var status string
	if cap.Status != PaypalPending {
		fmt.Errorf("not payapl pending order")
		return 0, data.Order{}, false
	}

	ID := cap.InvoiceID // our order id
	logger.Debugf("notified pending result is %v", cap)

	paymentid := getOrderIDFromUrl(cap.Links)

	databin, _ := json.Marshal(notifyData)
	pendingdetail, _ := json.Marshal(cap)

	TransactionId := cap.ID // pending order id is transaction id
	logger.Warnf("Transaction id is %s", TransactionId)
	order := data.Order{

		Status: status,
		//OrderRequest:  record.Request,
		PendingInfo:   string(pendingdetail),
		OrderID:       ID,
		PaymentID:     paymentid,
		TransactionID: TransactionId,
		PaymentVendor: "paypal",
		PaymentMethod: notifyData.ResourceType,
		PaymentNote:   brand_name,
		NotifyInfo:    string(databin[:]),
		//Description:   detail, //notifyData.Summary + "," + ID, // just for flil up incase no item list
		Currency:     cap.Amount.Currency,
		Total:        cap.Amount.Value,
		AdminNote:    "",
		UpdateTime:   fmt.Sprint(time.Now().Format(time.RFC1123)),
		Paytime:      fmt.Sprint(time.Now().Format(time.RFC1123)),
		IsRefund:     false,
		IsChargeBack: false,
	}

	request, err := data.GetRequestByID(ID)

	if err == nil {
		order.User = request.Email //.BuyerEmail
		order.PayerIP = request.IPAddr
		order.PayerLink = request.ContactInfo
		order.Comments = request.RequestInfo
		order.RequestTime = fmt.Sprint(time.Unix(request.OrderDate, 0).Format(time.RFC1123))
		purchaselist, _ := json.MarshalIndent(request.ItemList, "", "  ")
		order.Description = string(purchaselist)
		order.OrderDetail = order.Description
	} else {
		logger.Errorf("The request is not found, This is unusual,ID is %f", ID)
	}

	logger.Warn("Paypal Notified message is completed. checking exsiting order with order_id")
	// validreturn 0, data.Order{}, false

	order.Status = data.OrderPending //标识已付
	mm, _ := json.Marshal(order)

	retcode, ok := admin.CreateContent("Order", mm)

	if ok {
		logger.Debug("Pending Order created!", retcode)
		return retcode, order, true
	} else {
		return 0, data.Order{}, false
	}

}

// get order index id base on order_id
func GetIdByOrderID(id string) (string, error) {

	oid := admin.FindContentID("Order", id, "order_id")
	// update the record
	if len(oid) > 0 {
		return oid, nil
	}
	return "", fmt.Errorf("not found")
}

//get the order content return the data.Order structure data
func GetOrder(id string) (*data.Order, error) {
	orderbytes, err := db.Content("Order:" + id)
	if err != nil {
		return &data.Order{}, err
	}
	var order data.Order
	err = json.Unmarshal(orderbytes, &order)
	if err != nil {
		return &order, err
	}
	return &order, nil
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
		logger.Infof("Update Order %s status to %s is done !", id, state)
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

// update order with payment_id = id
func UpdateOrderByOrderID(id string, newOrder *data.Order) bool {
	// find paymenid is id record

	oid := admin.FindContentID("Order", id, "order_id")
	toUpdateData := url.Values{}
	/*
		ok := UpdateOrderStatusByID(oid, status)
				aa, err := admin.UpdateContent("Order", oid, "transaction_id", []byte(order.TransactionID))
				aa, err = admin.UpdateContent("Order", oid, "payer_link", []byte(order.PayerLink))
				aa, err = admin.UpdateContent("Order", oid, "notify_info", []byte(order.NotifyInfo))
				//net
				//Paid
				//paytime
	*/
	toUpdateData.Add("status", newOrder.Status)
	toUpdateData.Add("transaction_id", newOrder.TransactionID)
	toUpdateData.Add("payer_link", newOrder.PayerLink)
	toUpdateData.Add("payer", newOrder.Payer)
	toUpdateData.Add("notify_info", newOrder.NotifyInfo)
	toUpdateData.Add("net", newOrder.Net)
	toUpdateData.Add("paid", newOrder.Paid)
	toUpdateData.Add("pay_time", newOrder.Paytime)
	toUpdateData.Add("description", newOrder.Description)
	keys := []string{"status", "transaction_id", "payer_link", "payer", "notify_info", "net", "paid", "pay_time", "description"}
	_, err := admin.UpdateContents("Order", oid, keys, &toUpdateData)
	if err == nil {
		return true
	}
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

// The parameter id is order index id, so we need find the transactionID to search result

func GetTransationDetail(id string) (*SearchTransactionDetails, error) {
	logger.Debug("==================search transaction data  for id :", id)
	order, err := GetOrder(id)

	if err != nil { // no such order id
		return &SearchTransactionDetails{}, fmt.Errorf("order id %s is not found", id)
	}
	logger.Debugf("try to get order id %s for search transaction order id %s", order.OrderID, order.TransactionID)
	// first get data from ipn db
	notification, err := GetIPNbyID(order.OrderID)
	if err == nil { // have the ipn record
		logger.Debugf("notification get %s", notification)
		userstatus := "N"
		addrstatus := "N"
		if notification.PayerStatus == "verified" {
			userstatus = "Y"
		}
		if notification.AddressStatus == "confirmed" {
			addrstatus = "Y"
		}
		sdtail := SearchTransactionDetails{}
		sdtail.PayerInfo = SearchPayerInfo{
			PayerName: SearchPayerName{
				GivenName: notification.FirstName,
				Surname:   notification.LastName,
			},
			EmailAddress:  notification.PayerEmail,
			PayerStatus:   userstatus,
			AddressStatus: addrstatus,
		}
		sdtail.ShippingInfo = SearchShippingInfo{
			Name: notification.AddressName,
			Address: Address{
				Line1:       notification.AddressStreet,
				City:        notification.AddressCity,
				PostalCode:  notification.AddressZip,
				CountryCode: notification.AddressCountryCode,
			},
		}
		sdtail.TransactionInfo = SearchTransactionInfo{
			InvoiceID:              notification.Invoice,
			PayPalAccountID:        notification.PayerID,
			TransactionID:          notification.TxnID,
			TransactionSubject:     notification.TxnType,
			TransactionStatus:      string(notification.PaymentStatus),
			TransactionUpdatedDate: notification.PaymentDate, //.Time.Local().Format("yyyy-MM-dd HH:mm:ss"),
			PaymentMethodType:      string(notification.PaymentType),
			TransactionAmount: Money{
				Currency: notification.Currency,
				Value:    fmt.Sprintf("%.2f", notification.Gross),
			},
			FeeAmount: &Money{
				Currency: notification.Currency,
				Value:    fmt.Sprintf("%.2f", notification.Fee),
			},
		}
		return &sdtail, nil // fill enough data and return
	}
	logger.Error("We got ipn call with error:%s,have to turn to transaction search", err.Error())
	// Following the get transacation data from paypal and transaction db
	now := time.Now()
	now = now.Add(10 * time.Hour)
	last := now.AddDate(0, 0, -30) //30天的单子
	//fmt.Print(now, last)
	page := 1
	pageSize := 5
	allField := "all"

	//iid := order.TransactionID

	req := TransactionSearchRequest{
		//	TransactionID: &tid,
		EndDate:   now,
		StartDate: last,
		Page:      &page,
		PageSize:  &pageSize,
		Fields:    &allField,
	}
	if order.PaymentVendor != "paypal" {
		return &SearchTransactionDetails{}, fmt.Errorf("it is not paypal order, payment vendor is %s", order.PaymentVendor)
	}
	historytx, err := GetTransactionByID(order.OrderID)
	if err == nil {
		return historytx, err
	}

	req.TransactionID = &order.TransactionID

	result, err := payClient.ListTransactions(&req)
	if err == nil {
		//logger.Debug(result)

		if len(result.TransactionDetails) > 0 {

			SaveTransactions(&result.TransactionDetails[0])
			return &result.TransactionDetails[0], nil

		}
		return &SearchTransactionDetails{}, fmt.Errorf("no data from paypal")
	} else {
		logger.Debug(err)
		return &SearchTransactionDetails{}, err
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

func GetTransactionByID(orderid string) (*SearchTransactionDetails, error) {

	ret := make([]byte, 0)
	err := data.PaymentDBHandler.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(data.DBPayPayTransaction))

		ret = b.Get([]byte(orderid))

		return nil
	})
	if err != nil {
		logger.Error("Get order state error ", err)
		return nil, err
	}
	rr := SearchTransactionDetails{}

	datasrc := bytes.NewReader(ret)
	err = json.NewDecoder(datasrc).Decode(&rr)
	return &rr, err
}

// SaveRequest to save the data in payment
func SaveTransactions(r *SearchTransactionDetails) error {

	if r == nil {
		return fmt.Errorf("No transaction info data,save transaction error")
	}
	ID := r.TransactionInfo.InvoiceID
	_, err := GetTransactionByID(ID)
	if err == nil {
		return fmt.Errorf("the transcation exists")
	}
	tx, err := data.PaymentDBHandler.Begin(true)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer tx.Rollback()

	root := tx.Bucket([]byte(data.DBPayPayTransaction))

	if buf, err := json.Marshal(r); err != nil {
		logger.Error(err)
		return data.PaymentErrInputData
	} else if err := root.Put([]byte(ID), buf); err != nil {
		logger.Error(err)
		return data.PaymentErrSave
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return data.PaymentErrSave
	}

	logger.Info("Save user transaction search result. ")

	return nil
}

func GetIPNbyID(orderid string) (*IPNNotification, error) {

	ret := make([]byte, 0)
	err := data.PaymentDBHandler.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(data.DBPayPayIPN))

		ret = b.Get([]byte(orderid))

		return nil
	})
	if err != nil {
		logger.Error("Get order state error ", err)
		return nil, err
	}
	rr := IPNNotification{}
	logger.Debugf(string(ret))
	datasrc := bytes.NewReader(ret)
	err = json.NewDecoder(datasrc).Decode(&rr)
	logger.Debug(err)
	return &rr, err
}

// here save the ipn data into db for later search
func SaveIPNData(notification *IPNNotification) error {

	if notification == nil {
		return fmt.Errorf("No ipn info data,save error")
	}
	ID := notification.Invoice
	_, err := GetTransactionByID(ID)
	if err == nil {
		logger.Warnf("the transcation exists:%v", ID)
		return nil
	}
	tx, err := data.PaymentDBHandler.Begin(true)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer tx.Rollback()

	root := tx.Bucket([]byte(data.DBPayPayIPN))

	if buf, err := json.Marshal(notification); err != nil {
		logger.Error(err)
		return data.PaymentErrInputData
	} else if err := root.Put([]byte(ID), buf); err != nil {
		logger.Error(err)
		return data.PaymentErrIPNSave
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return data.PaymentErrIPNSave
	}

	logger.Info("ipn data Saved.", ID)

	return nil
}
