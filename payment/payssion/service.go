package payssion

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/agreyfox/eshop/payment/data"
)

const (
	brand_name string = "egpal 公司"
	prefer     string = "return=representation"
)

var (
	returnURL = "https://support.bk.cloudns.cc:8081/payment/payssion/return"
	cancelURL = "https://support.bk.cloudns.cc:8081/payment/payssion/cancle"

	//APIKey    = "90a00a8dc3231897"
	//SecretKey = "0f0772dc61a1480c2fe80f9a4e1b2c85"
	APIKey    = "sandbox_5dea43e2a2a8e257"
	SecretKey = "nEtnfCWfLgc5GuiEobU8hvfp7O4CPR0c"

	notifyURL = "http://www.egpal.com/payssion_notify.html"
	emailURL  = "aocdepot@gmail.com"
	payClient *Client
	payMethod string
)

func initPayssion() {

	client, _ := NewClient(APIKey, SecretKey, APIBaseSandBox)
	client.SetLog(logger) // Set log to terminal stdout
	payClient = client

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
	logger.Debug("User create the payssion payment")
	ip := GetIP(r)
	//try to get user post information about the payment

	reqJSON := getJSONFromBody(r)

	pmid, ok := reqJSON["pm_id"].(string)
	if !ok || len(pmid) == 0 {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "pm_id error",
		})
		return
	}
	if !isRightPMID(pmid) {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "wrong pm_id",
		})
		return
	}

	amount, ok := reqJSON["amount"].(string)
	if !ok || len(amount) == 0 {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "amount error  ",
		})
		return
	}
	currency, ok := reqJSON["currency"].(string)
	if !ok || len(currency) == 0 {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "currency error  ",
		})
		return
	}
	desc, ok := reqJSON["description"].(string)
	if !ok {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "description error  ",
		})
		return
	}
	orderid, ok := reqJSON["order_id"].(string)
	if !ok { //若没有orderid,生成一个
		orderid = getOrderID()
	}

	sigArray := []string{APIKey, pmid, amount, currency, orderid, SecretKey}
	appsig := Signature(sigArray)

	email, ok := reqJSON["payer_email"].(string)
	if !ok {
		email = ""
	}
	payname, ok := reqJSON["payer_name"].(string)
	if !ok {
		payname = ""
	}

	purchaseReq := PaymentRequest{
		APIKey:      APIKey,
		PMID:        pmid,
		Amount:      amount,
		Currency:    currency,
		Description: desc,
		OrderID:     orderid,
		APISig:      appsig,
		//ReturnURL:   returnURL,
		PayerEmail: email,
		PayerName:  payname,
		PayerIP:    ip,
	}

	logger.Debug("Ready to create a payssion order,Data Save to tempdb!")

	//	client.Lock()
	order, err := payClient.CreateOrder(&purchaseReq)
	//	client.Unlock()
	if err != nil {
		logger.Debug("create payssion order error:", err)
		renderJSON(w, r, map[string]interface{}{
			"retCode": -2,
			"msg":     "创建订单失败",
		})
		return
	}
	go saveRequest(order, reqJSON)

	if order.ResultCode == PayssionOK {

		renderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    order,
		})
	} else {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -10,
			"msg":     "Something error ",
			"data":    order.Transaction.State,
		})
	}

}

func excutePayment(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User excute the payment")
}

/*
respond to paypal redirection
https://view.bk.cloudns.cc:8080/payment/payssion/return?transaction_id=T522221218470524&order_id=1234

https://ssl.dotpay.pl/test_payment/result/M9962-98611/a8a063a28e5755a862ed71ba140ba29e69d68d4f75cf0aff570468f25943446d/
http://view.bk.cloudns.cc:8080/payment/payssion/return?transaction_id=T525675987024141&order_id=1234

*/

func Succeed(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Payssion return  data from payment")
	q := r.URL.Query()

	payID := q.Get("transaction_id")
	order_id := q.Get("order_id")
	logger.Debugf("User finished payment %s is down  order id  is %s,", payID, order_id)

	//w.Write([]byte(fmt.Sprintf("订单号 ：%s已付款,Thanks！", order_id)))
	url := data.OnlineURL
	if len(order_id) > 0 {
		url += fmt.Sprintf("?status=1&orderno=%s", order_id)
	} else {
		url += fmt.Sprintf("status=0&msg=%s", "Payment finished with error")
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func Notify(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Get Notify data from payssion")

	//fmt.Println("got you")
	re := getJSONFromBody(r)

	orderState := &NotifyResponse{}
	bc, err := json.Marshal(re)
	if err == nil {
		err = json.Unmarshal(bc, orderState)
		go saveNotify(orderState)
		if err == nil {
			sigArray := []string{APIKey, orderState.PMID, orderState.Amount, orderState.Currency, orderState.OrderID, orderState.State, SecretKey}
			verify := Signature(sigArray)
			if orderState.NotifySignature == verify {
				switch orderState.State {
				case PayssionCompleted:
					logger.Warnf("Order Completed! App is %s,Transactions id:%s,Total:%s ", orderState.AppName, orderState.TransactionID, orderState.Amount)
					// notification verify string: api_key|pm_id|amount|currency|order_id|state|sercret_key
					logger.Debug("Signature is %s", verify)

					oid, oo, ok := CreateNewOrderInDB(orderState)
					if ok {
						logger.Infof("Order %d created! and orderid %s ", oid, orderState.OrderID)
						if len(oo.Payer) > 0 {
							go data.SendConfirmEmail(oo.OrderID, orderState.AppName, orderState.Amount, orderState.Currency, oo.Payer)
						}
					} else {
						logger.Warn("Create %s order error,Need check!", orderState.OrderID)
					}
				case PayssionChargeBack:
					logger.Warnf("The order %s state had chargeback  to %s,total:%s", orderState.TransactionID, orderState.Amount)
					UpdateOrderStatus(orderState.TransactionID, orderState.OrderID, orderState.State)
				case PayssionCreated:
					logger.Infof("The transactions %s is be created order id is %s,total amount is  %s", orderState.TransactionID, orderState.OrderID, orderState.Amount)
				case PayssionRefund:
					logger.Warnf("Refunded !,Order id is %s, Amount:%s", orderState.OrderID, orderState.Amount)
					UpdateOrderStatus(orderState.TransactionID, orderState.OrderID, orderState.State)
				default:
					logger.Debugf("Order %s state change to %s", orderState.OrderID, orderState.State)
				}

			} else {
				logger.Warnf("Wrong signature for order %s:", orderState.OrderID)
			}

			w.WriteHeader(http.StatusOK)
			return
		} else {
			logger.Error("Notify with parse error ", err)
		}
	} else {
		logger.Error("Error  in notification data ", err)
	}
	w.WriteHeader(http.StatusBadRequest) // to be carefuly this return to payssion code. need check
}

func Failed(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User Cancle the payment")

}

func Index(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Index open!")
	w.Write([]byte("PayPal Index"))
}
