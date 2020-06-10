package skrill

import (
	"errors"
	"fmt"

	"net/http"
	"net/url"
	"strconv"
)

var (
	NotifyURL      = "http://view.bk.cloudns.cc:8080/payment/skrill/notify"
	ReturnURL      = "http://view.bk.cloudns.cc:8080/payment/skrill/return"
	CancelURL      = "http://view.bk.cloudns.cc:8080/payment/skrill/cancel"
	TransactionID  = ""
	MerchanAccount = ""
	payClient      *Client
	payMethod      string
)

func initSkrill() {
	logger.Debug("Skrill backend service initialized!")
	payClient = New()
}

func CreateTest() {
	payClient = New()
	para := PrepareParam{
		Amount:    3.4,
		Currency:  EUR,
		StatusURL: ReturnURL,
	}
	payClient.Prepare(&para)
}

// When user to checkout "Pay Now" button ,It will send the request to beckend system and beckend system will
// send the request to create the payment.return the created payment information with authorization url
/* input data is looks like

Name	Type	Required	Description
api_key	string	Yes	The api key of your app  XX

pm_id	string	Yes	The payment method id: e.g. alipay_cn. See the pm_id list list for details
amount	string	Yes	The payment amount: e.g. 10.12
currency	string	Yes	3-letter ISO code for currency: e.g. USD
description	string	Yes	payment description: e.g. example.com #item.
order_id	string	Yes	The order id for this payment
api_sig	string	Yes	The api request signature. See how to generate the signature
return_url	string	Optional	The URL the customer should be redirected to after the payment no matter if the customer has completed the payment. You need to set the default return URL in your app if leaving it blank in the request
payer_email	string	Optional	The customer’s email
payer_name	string	Optional	The customer’s name
*/
func createPayment(w http.ResponseWriter, r *http.Request) {

	ip := GetIP(r)
	//try to get user post information about the payment
	logger.Debugf("User create the paypal payment from %s", ip)
	reqJSON := getJSONFromBody(r)
	reqJSON["ip"] = GetIP(r)
	logger.Debug("User request is %s", reqJSON)

	s, err := strconv.ParseFloat(fmt.Sprintf("%s", reqJSON["amount"]), 64)
	if err != nil {
		logger.Error("Amount data error ")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c := fmt.Sprintf("%s,", reqJSON["currency"])

	para := PrepareParam{
		Amount:             s,
		Currency:           GetCurrencyCode(c),
		ReturnURL:          ReturnURL,
		StatusURL:          NotifyURL,
		Language:           Language(fmt.Sprintf("%s", reqJSON["language"])),
		LogoURL:            fmt.Sprintf("%s", reqJSON["logo_url"]),
		PayFromEmail:       fmt.Sprintf("%s", reqJSON["payer_email"]),
		MerchantFields:     "order_id",
		OrderID:            getOrderID(),
		Address:            fmt.Sprintf("%s", reqJSON["address"]),
		PhoneNumber:        fmt.Sprintf("%s", reqJSON["phone_number"]),
		City:               fmt.Sprintf("%s", reqJSON["city"]),
		Country:            fmt.Sprintf("%s", reqJSON["country"]),
		Detail1Description: fmt.Sprintf("%s", reqJSON["description"]),
		FirstName:          fmt.Sprintf("%s", reqJSON["firstname"]),
		LastName:           fmt.Sprintf("%s", reqJSON["lastname"]),
		PaymentMethods:     fmt.Sprintf("%s", reqJSON["payment_methods"]),
	}
	ll, err := payClient.Prepare(&para)

	if err == nil {
		para.TransactionID = ll
		retData := map[string]interface{}{
			"transaction":  para,
			"redirect_url": ll,
		}
		go saveRequest(&para, reqJSON)
		renderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    retData,
		})
	} else {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -10,
			"msg":     "Something error ",
			"data":    fmt.Sprint(err),
		})
	}
	w.WriteHeader(http.StatusLocked)
}

func excutePayment(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User excute the payment")
}

func Succeed(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Skrill return  data from payment")
	//q := r.URL.Query()
	///logger.Debugf("%v\n%s", r.URL, q)
	//data := getJSONFromBody(r)
	bodybytes := getBinaryDataFromBody(r)
	logger.Debugf("User finished payment with data %v,", string(bodybytes[:]))

	w.Write([]byte("订单号 ：已付款,Thanks！"))

}

func SaveNotify(data []byte) error {
	values, err := url.ParseQuery(string(data))
	logger.Debug(values)
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
			return errors.New("State parse  error ")
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

		if retData.Status == SkrillProcessed { // create ok
			CreateNewOrderInDB(&retData)
			return nil
		} else {
			logger.Debugf("Notify data with status %s", Status(statusvalue))
			return errors.New("Notify result is not OK:" + Status(statusvalue).String())
		}
	} else {
		logger.Error("Parse Body data error ", err)
		return errors.New("Wrong input data")
	}

	return nil
}

func Notify(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Get Notify data from Skrill:", GetIP(r))
	bodybytes := getBinaryDataFromBody(r)
	logger.Debug(string(bodybytes[:]))
	/*
		transaction_id=3195856960
		&mb_amount=1.3
		&amount=1.3
		&md5sig=B23743880D2FAE5D02F0205ABBF9B6FA
		&merchant_id=138853317
		&payment_type=WLT
		&mb_transaction_id=3195856960
		&mb_currency=USD
		&pay_from_email=18901882538%40189.cn
		&pay_to_email=e_raeb%40163.com
		&currency=USD
		&customer_id=139601073
		&status=2

	*/
	err := SaveNotify(bodybytes)
	if err != nil {
		logger.Error("Process data error ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK) // to be carefuly this return to payssion code. need check
}

func Failed(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User Cancel the payment")
	logger.Debug(r)
	bbb := getBinaryDataFromBody(r)
	logger.Debug(string(bbb[:]))

}

func Index(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Index open!")
	w.Write([]byte("Skrill Index"))
}
