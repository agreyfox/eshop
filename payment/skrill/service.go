package skrill

import (
	"encoding/json"
	"fmt"
	"time"

	"net/http"
	"strconv"

	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/prometheus"
)

var (
	TransactionID  = ""
	MerchanAccount = ""
	payClient      *Client
	payMethod      string
)

func CreateTest() {
	payClient = New()
	para := PrepareParam{
		Amount:    3.4,
		Currency:  EUR,
		StatusURL: ReturnURL,
	}
	payClient.Prepare(&para)
}

func userSubmit(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User submit a  payssion  payment")
	//ip := GetIP(r)
	//try to get user post information about the payment

	payload := new(data.UserSubmitOrderRequest)
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		//RespondError(w, err, http.StatusBadRequest, GetReqID(r))
		logger.Errorf("user submit error", err)
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "input data parse error",
		})
		return
	}
	//reqJSON := getJSONFromBody(r)
	payload.IPAddr = data.GetIP(r)
	if validateRequest(payload) != nil {
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "input data parse error",
		})
		return
	}
	payload.OrderID = data.GetShortOrderID()
	payload.OrderDate = time.Now().Unix()
	respond, err := createOrder(payload) //create payssion call

	payload.Respond = respond

	errsave := data.SaveOrderRequest(payload) //finished save request,
	logger.Debug(errsave)
	if err == nil {

		retData := map[string]interface{}{
			"transaction":  payload,
			"redirect_url": respond,
		}
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    retData,
		})

	} else {
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -10,
			"msg":     "Something error",
			"data":    fmt.Sprint(err),
		})
	}
	w.WriteHeader(http.StatusLocked)
	return

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

	ip := data.GetIP(r)
	//try to get user post information about the payment
	logger.Debugf("User create the paypal payment from %s", ip)
	reqJSON := data.GetJSONFromBody(r)
	reqJSON["ip"] = ip
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
		OrderID:            data.GetShortOrderID(),
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

		go prometheus.OrderCounter.WithLabelValues("skirll").Add(1) //metric order creation

		para.TransactionID = ll
		retData := map[string]interface{}{
			"transaction":  para,
			"redirect_url": ll,
		}
		go saveRequest(&para, reqJSON)
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    retData,
		})
	} else {
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -10,
			"msg":     "Something error",
			"data":    fmt.Sprint(err),
		})
	}
	w.WriteHeader(http.StatusLocked)
}

func excutePayment(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User excute the payment")
}

func Succeed(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	logger.Infof("Skrill return data transaction_id is %s , msid is %s\n", q.Get("transaction_id"), q.Get("msid"))

	logger.Debugf("Notify from skril with request:%v\n%v", r, r.URL)

	url := data.OnlineURL

	tid := q.Get("transaction_id")
	if len(tid) == 0 {
		logger.Error("skrill system return strange data, please check skrill account for transaction ")
		url += fmt.Sprint("?status=0&msg=skrill system return strange data, please check skrill account for transaction")
	} else {
		//w.Write([]byte("订单号 ：已付款,Thanks！"))
		//order_id := fmt.Sprint(target["order_id"])

		url += fmt.Sprintf("?status=1&orderno=%s", tid)

	}

	http.Redirect(w, r, url, http.StatusFound)
}

func Notify(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Get Notify data from Skrill:", data.GetIP(r))
	bodybytes := data.GetBinaryDataFromBody(r)
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

	err := CreateOrderByNotify(bodybytes)
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
	bbb := data.GetBinaryDataFromBody(r)
	logger.Debug(string(bbb[:]))

}

func Index(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Index open!")
	w.Write([]byte("Skrill Index"))
}
