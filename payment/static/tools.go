package static

import (
	"encoding/json"
	"fmt"

	"github.com/agreyfox/eshop/payment/data"

	"time"

	"github.com/agreyfox/eshop/system/admin"
)

// valid user request for payssion
func validateRequest(req *data.UserSubmitOrderRequest) error {

	if req.Amount <= 0 {
		return fmt.Errorf("amount data error")
	}
	if req.Amount < 0.0001 {
		return fmt.Errorf("amount data too small") //w.Write([]byte("订单号 ：已付款,Thanks！"))
		//order_id := fmt.Sprint(target["order_id"])
		return fmt.Errorf("no payer id")
	}
	if len(req.RequestInfo) == 0 {
		return fmt.Errorf("no user request info")
	}
	if len(req.Currency) == 0 {
		return fmt.Errorf("no user currency info")
	}
	if len(req.Payment) == 0 {
		return fmt.Errorf("no payment")
	}
	if len(req.PaymentChannel) == 0 {
		return fmt.Errorf("no payment channel info")
	}
	return nil
}

// to save a record
func saveRequest(r *data.UserSubmitOrderRequest) bool {
	// Store the user model in the user bucket using the username as the key.
	da, _ := json.Marshal(r)

	record := data.PaymentLog{
		RequestData: string(da),
	}
	record.OrderID = r.OrderID
	record.PaymentMethod = "static"
	record.PaymentID = r.OrderID
	record.Total = fmt.Sprintf("%.2f", r.Amount)
	record.Currency = fmt.Sprintf("%s", r.Currency)
	record.PaymentState = data.OrderCreated // use crated as status
	record.BuyerEmail = r.Email

	record.RequestTime = time.Now().Unix()

	record.Comments = r.RequestInfo

	record.IP = r.IPAddr

	ok := data.SavePaymentLog(&record)
	if ok {
		logger.Info("Save payment log done!") //w.Write([]byte("订单号 ：已付款,Thanks！"))

	}
	return ok
}

// create static order
func createOrder(r *data.UserSubmitOrderRequest) (string, error) {

	logger.Infof("User create the static page  payment from %s", r.IPAddr)
	data.SaveOrderRequest(r)
	ss, oo, ok := CreateNewOrderByRequest(r)
	fmt.Println(ss, oo, ok)
	return payClient.GetReturnPage(), nil

}

//直接根据request创建定单
func CreateNewOrderByRequest(reqdata *data.UserSubmitOrderRequest) (int, data.Order, bool) {

	oid := reqdata.OrderID
	capdetail, _ := json.Marshal(reqdata.ItemList)
	detail := string(capdetail)

	buff, _ := json.Marshal(reqdata)
	order := data.Order{

		Status:        data.OrderCreated, //创建订单
		OrderID:       oid,
		PaymentVendor: "static",
		PaymentMethod: reqdata.PaymentChannel,
		PaymentNote:   "",
		NotifyInfo:    string(buff[:]),
		Description:   detail,
		Total:         fmt.Sprintf("%.2f", reqdata.Amount),
		Paid:          "",
		Net:           "",
		AdminNote:     "", //oid + fmt.Sprintf("Return code is %s", notifyData.FailedReasonCode),
		UpdateTime:    fmt.Sprint(time.Now().Format(time.RFC1123)),
		Paytime:       fmt.Sprint(time.Now().Format(time.RFC1123)),
		IsRefund:      false,
		IsChargeBack:  false,
		User:          reqdata.Email,
	}

	databuff, _ := json.Marshal(reqdata.ItemList)
	order.OrderDetail = string(databuff)
	order.Payer = ""
	order.PayerLink = reqdata.ContactInfo
	order.Comments = reqdata.RequestInfo
	order.PayerIP = reqdata.IPAddr
	order.RequestTime = fmt.Sprint(time.Unix(reqdata.OrderDate, 0).Format(time.RFC1123))

	///order.Status = data.OrderPaid // create a order start from order paid
	mm, _ := json.Marshal(order)

	retCode, ok := admin.CreateContent("Order", mm)

	if ok {
		return retCode, order, true
	} else {
		logger.Error("error in creat order with code :", retCode)
		return 0, order, false
	}

}
