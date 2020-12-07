package payssion

import (
	"encoding/json"
	"fmt"

	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/system/admin"

	"time"
)

// valid user request for payssion
func validateRequest(req *data.UserSubmitOrderRequest) error {

	if !isRightPMID(req.PaymentChannel) {
		return fmt.Errorf("wrong pm_id")
	}

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

func createOrder(r *data.UserSubmitOrderRequest) (*PaymentResponse, error) {
	pmid := r.PaymentChannel
	pmid, amount, currency, orderid := r.PaymentChannel, r.Amount, r.Currency, r.OrderID
	//sigArray := []string{APIKey, pmid, amount, currency, orderid, SecretKey}
	sigArray := []string{APIKey, pmid, fmt.Sprint(amount), currency, orderid, SecretKey}
	appsig := Signature(sigArray)

	email, payname, desc := r.Email, r.LastName+" "+r.FirstName, r.RequestInfo

	purchaseReq := PaymentRequest{
		APIKey:      APIKey,
		PMID:        pmid,
		Amount:      fmt.Sprint(amount),
		Currency:    currency,
		Description: desc,
		OrderID:     orderid,
		APISig:      appsig,
		//ReturnURL:   returnURL,
		PayerEmail: email,
		PayerName:  payname,
		PayerIP:    r.IPAddr,
	}

	logger.Debug("Ready to create a payssion order,Data Save to tempdb!")
	//logger.Debug(purchaseReq)
	//	client.Lock()
	order, err := payClient.CreateOrder(&purchaseReq)
	//	client.Unlock()
	if err != nil {
		logger.Errorf("create payssion order error:%s", err)
		return nil, fmt.Errorf("创建订单失败:%s", err.Error())
	}
	//go saveRequest(order, reqJSON)

	if order.ResultCode == PayssionOK {
		return order, nil
		/* renderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    order,
		}) */
	} else {
		return order, fmt.Errorf("创建失败:%s", order.Transaction.State)
		/* 	renderJSON(w, r, map[string]interface{}{
			"retCode": -10,
			"msg":     "Something error ",
			"data":    order.Transaction.State,
		})
		*/
	}

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
	record.OrderID = t.OrderID + "-" + t.State // log id format
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

// CreateNewOrderInDB try to create new order in backend system with the information provide data and the record
/*
 system db structure is
	orderid:order structure
*/
func CreateNewOrderInDB(notifyData *NotifyResponse) (int, data.Order, bool) {

	oid := notifyData.OrderID
	ID := oid
	//val := &bytes.Buffer{}
	//states, err := data.GetLog(ID)
	//logger.Info(states)  // remove 2020/11/05
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
		//Comments:      notifyData.Description,
		Currency:     notifyData.Currency,
		Total:        notifyData.Amount,
		Paid:         notifyData.Paid,
		Net:          notifyData.Net,
		AdminNote:    oid,
		UpdateTime:   fmt.Sprint(time.Now().Format(time.RFC1123)),
		Paytime:      fmt.Sprint(time.Now().Format(time.RFC1123)),
		IsRefund:     false,
		IsChargeBack: false,
	}
	// now to get some data from payment request

	request, err := data.GetRequestByID(ID)
	/* 	originReq, err := data.GetRequestByState(ID, PayssionPending)

	   	if err == nil {
	   		databuff, _ := json.Marshal(originReq.ReturnData)
	   		order.OrderDetail = string(databuff[:])
	   		order.Payer = originReq.BuyerEmail
	   		order.PayerIP = originReq.IP
	   		order.RequestTime = fmt.Sprint(time.Unix(originReq.RequestTime, 0).Format(time.RFC1123))
	   	} */
	if err == nil {
		databuff, _ := json.MarshalIndent(request.ItemList, "", "  ")
		order.OrderDetail = string(databuff[:])
		order.User = request.Email
		order.PayerLink = request.ContactInfo
		order.PayerIP = request.IPAddr
		order.Payer = ""
		order.Comments = request.RequestInfo
		order.RequestTime = fmt.Sprint(time.Unix(request.OrderDate, 0).Format(time.RFC1123))
	} else {
		logger.Errorf("The request is not found, This is wired!,ID is %f", ID)
	}
	order.Status = data.OrderPaid // create a order start from order paid
	mm, _ := json.Marshal(order)

	retCode, ok := admin.CreateContent("Order", mm)

	//admin.UpdateContent("Order", fmt.Sprintf("%d", retCode), "status", []byte(data.OrderInValidate))
	//to update the order status
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
