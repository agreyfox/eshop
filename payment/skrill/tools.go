package skrill

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

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

// to save a record
func saveRequest(t *PrepareParam, request map[string]interface{}) bool {
	// Store the user model in the user bucket using the username as the key.

	record := data.PaymentLog{
		RequestData: request,
		ReturnData:  t,
	}
	record.OrderID = t.OrderID
	record.PaymentMethod = "skrill"
	record.PaymentID = t.TransactionID
	record.Total = fmt.Sprintf("%.2f", t.Amount)
	record.Currency = fmt.Sprintf("%s", t.Currency)
	record.PaymentState = SkrillCreated.String() // use crated as status
	record.BuyerEmail = t.PayFromEmail

	record.RequestTime = time.Now().Unix()

	record.Comments = t.Detail1Description

	record.IP = fmt.Sprintf("%s", request["ip"])

	ok := data.SavePaymentLog(&record)
	if ok {
		logger.Info("Save payment log done!")
	} else {
		logger.Info("Save payment log error!")
	}
	return ok

}

// to save a record
func saveNotify(t *StatusResponse) bool {
	// Store the user model in the user bucket using the username as the key.

	record := data.PaymentLog{
		ReturnData: t,
	}
	record.OrderID = t.OrderID + "-" + t.Status.String() // log id format
	record.PaymentMethod = "payssion"
	record.PaymentID = t.MbTransactionID
	record.Total = fmt.Sprint(t.MbAmount)
	record.Currency = fmt.Sprint(t.MbCurrency)
	record.PaymentState = t.Status.String()

	record.RequestTime = time.Now().Unix()
	record.Info = t.PaymentType

	ok := data.SavePaymentLog(&record)
	if ok {
		logger.Info("Save skrill payment log done!")
	} else {
		logger.Info("Save skrill payment log error!")
	}
	return ok

}

// create skril
func createOrder(r *data.UserSubmitOrderRequest) (string, error) {

	logger.Infof("User create the Skrill payment from %s", r.IPAddr)

	s := r.Amount
	c := r.Currency

	para := PrepareParam{
		Amount:             s,
		Currency:           GetCurrencyCode(c),
		ReturnURL:          ReturnURL,
		StatusURL:          NotifyURL,
		Language:           Language(fmt.Sprintf("%s", r.Language)),
		LogoURL:            fmt.Sprintf("%s", r.LogoURL),
		PayFromEmail:       fmt.Sprintf("%s", r.Email),
		MerchantFields:     "order_id",
		OrderID:            data.GetShortOrderID(),
		Address:            r.Address,
		PhoneNumber:        r.Phone,
		City:               r.City,           //fmt.Sprintf("%s", reqJSON["city"]),
		Country:            r.Country,        //.Sprintf("%s", reqJSON["country"]),
		Detail1Description: r.RequestInfo,    //fmt.Sprintf("%s", reqJSON["description"]),
		FirstName:          r.FirstName,      //fmt.Sprintf("%s", reqJSON["firstname"]),
		LastName:           r.LastName,       //fmt.Sprintf("%s", reqJSON["lastname"]),
		PaymentMethods:     r.PaymentChannel, //fmt.Sprintf("%s", reqJSON["payment_methods"]),
	}
	//logger.Debug(para)
	ll, err := payClient.Prepare(&para)
	return ll, err

}

// save notity to log, create order in dv
func CreateOrderByNotify(Notifydata []byte) error {
	values, err := url.ParseQuery(string(Notifydata))
	logger.Debugf("Get skril notify data :%v\n ready to create order", values)
	if err == nil {
		amm, err := strconv.ParseFloat(values.Get("amount"), 64)
		if err != nil {
			logger.Error("amount parse error ")
			return errors.New("amount parse error ")
		}
		mbmm, err := strconv.ParseFloat(values.Get("mb_amount"), 64)
		if err != nil {
			logger.Warn("mb amount parse error ")
		}
		statusvalue, err := strconv.ParseInt(values.Get("status"), 10, 0)
		if err != nil {
			logger.Error("Status parse error ")
			return errors.New("State parse  error")
		}
		failedReasonCode, err := strconv.ParseInt(values.Get("failed_reason_code"), 10, 0)
		if err != nil {
			logger.Warn("FailedReasonCode parse error ")
		}
		retData := StatusResponse{
			PayToEmail: values.Get("pay_to_email"),

			PayFromEmail: values.Get("pay_from_email"),

			MerchantID:       values.Get("merchant_id"),
			CustomerID:       values.Get("customer_id"),
			TransactionID:    values.Get("transaction_id"),
			MbTransactionID:  values.Get("mb_transaction_id"),
			MbAmount:         mbmm,
			MbCurrency:       GetCurrencyCode(values.Get("mb_currency")),
			Status:           Status(statusvalue),
			FailedReasonCode: Code(failedReasonCode),
			Md5Sig:           values.Get("md5sig"),
			Sha2Sig:          values.Get("sha2sig"),
			Amount:           amm,
			Currency:         Currency(values.Get("currency")),
			NetellerID:       values.Get("neteller_id"),
			PaymentType:      values.Get("payment_type"),
			OrderID:          values.Get("order_id"),
		}
		go saveNotify(&retData)                //add 2020/11/25
		if retData.Status == SkrillProcessed { // create ok
			oid, oo, ok := CreateNewOrderInDB(&retData)
			if ok {
				if len(oo.Payer) > 0 {
					go data.SendConfirmEmail(oo.OrderID, "Game item", values.Get("amount"), values.Get("currency"), oo.Payer)
				}
			}
			logger.Debugf("Skrill Order created with id:%d, OrderID is %s", oid, oo.OrderID)
			return nil
		} else {
			logger.Debugf("Notify data with status %s", Status(statusvalue))
			return errors.New("Notify result is not OK:" + Status(statusvalue).String())
		}
	} else {
		logger.Error("Parse Body data error ", err)
		return errors.New("wrong input data")
	}

	return nil
}

// CreateNewOrderInDB try to create new order in backend system with the information provide data and the record
/*
 system db structure is
	orderid:order structure
*/
func CreateNewOrderInDB(notifyData *StatusResponse) (int, data.Order, bool) {

	oid := notifyData.OrderID
	ID := oid

	buff, _ := json.Marshal(notifyData)
	order := data.Order{

		Status:        data.OrderPaid, //已支付
		OrderID:       oid,
		PaymentVendor: "skrill",
		PaymentMethod: notifyData.PaymentType,
		PaymentNote:   notifyData.NetellerID,
		NotifyInfo:    string(buff[:]),
		Currency:      string(notifyData.Currency),
		Total:         fmt.Sprintf("%.2f", notifyData.Amount),
		Paid:          "",
		Net:           "",
		AdminNote:     "",
		UpdateTime:    fmt.Sprint(time.Now().Format(time.RFC1123)),
		Paytime:       fmt.Sprint(time.Now().Format(time.RFC1123)),
		IsRefund:      false,
		IsChargeBack:  false,
		Payer:         notifyData.PayFromEmail,
	}

	request, err := data.GetRequestByID(ID)

	if err == nil {
		databuff, _ := json.Marshal(request.ItemList)
		order.OrderDetail = string(databuff[:])
		order.Description = string(databuff[:])
		order.User = request.Email
		order.PayerLink = request.ContactInfo
		order.PayerIP = request.IPAddr
		order.Comments = request.RequestInfo
		order.RequestTime = fmt.Sprint(time.Unix(request.OrderDate, 0).Format(time.RFC1123))
	} else {
		logger.Errorf("The request is not found, This is wired!,ID is %f", ID)
	}
	order.Status = data.OrderPaid // create a order start from order paid
	mm, _ := json.Marshal(order)

	retCode, ok := admin.CreateContent("Order", mm)

	if ok {
		return retCode, order, true
	} else {
		logger.Error("error in creat order with code :", retCode)
		return 0, order, false
	}

}
