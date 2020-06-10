package paypal

import (
	"encoding/json"
	"fmt"
	"github.com/agreyfox/eshop/payment/data"

	"github.com/robfig/cron"
	"net/http"
	"time"
)

const (
	brand_name string = "Abc 公司"
	prefer     string = "return=representation"
)

var (
	// paypal ClientID
	ClientID = "AbOMcM4iaf0PYKGgOFCktDD-Rqzpn7R_r2yPfwbopgCLYkBLXkD45c1qejwVX2BrBSxVQgz3_QlU7iFn"
	// Paypal client secrte
	Secret      = "EKxToL0apcJ7HOAryLeFkyP9JRWuw-p8pMj9M5N3Y1Ee8tsUDFgRv1wA_3hIjRMiHqrmbQu12KW_Noys"
	accessToken *TokenResponse
	returnURL   = "http://view.bk.cloudns.cc:8080/payment/paypal/return"
	cancelURL   = "http://view.bk.cloudns.cc:8080/payment/paypal/cancle"
	thanksURL   = "http://view.bk.cloudns.cc:8080/thanks"
	webHookID   = "3S402195M8327334K"
	hookURL     = "https://view.bk.cloudns.cc/payment/paypal/notify"
	payClient   *Client
	ProgramMode = "DEBUG"
)

func initpaypal() {

	client, err := NewClient(ClientID, Secret, APIBaseSandBox)
	client.SetLog(logger) // Set log to terminal stdout
	payClient = client
	logger.Debug("Paypal get access token result:", err)
	accessToken, err = payClient.GetAccessToken()
	logger.Debug(accessToken)

	logger.Info("Start Paypal access token refresh ")
	backendjob := cron.New()

	backendjob.AddFunc("@every 8h30m", func() {
		logger.Debug("paypal refresh token service")
		Refresh()
	})
	backendjob.Start()
}

// Refresh to 重新获取access token
func Refresh() {
	n := time.Now()
	if payClient.tokenExpiresAt.Sub(n).Seconds() < 3600 {
		logger.Debug("Update paypal accessToken ....")
		payClient.GetAccessToken()
		logger.Debugf("Update job done! Next will be:%s", payClient.tokenExpiresAt.Format(time.RFC3339))
	}

}

// When user to checkout "Pay Now" button ,It will send the request to beckend system and beckend system will
// send the request to create the payment.return the created payment information with authorization url
/* input data is looks like
{
	email:xxxxx   <=====购买者地址
	item_list:[
			{     <========第一个销售对象
			"reference_id":"abdc-1",
			"amount":{
					"currency_code":"USD",
					"value":"12",
					breakdown:{
						"item_total":{     <== must if we have detail items
							"currency_code":"USD",
							"value":"12",
						},
						"shipping":{},
						"handling":{},
						"tax_total":{},
						"insurance":{},
						"shipping_discount":{},
						"discount":{}
					}
			},
			"items":[
				{
					"name":"item name",
					"quantity":22,
					"unit_amount":{
						"currency_code":"USD",
						"value":"12"
					},
					"description":"user input for game item requirment"
				}
			],
			"invoice_id" :"customer invoiceid is herer",
			"customer_id":"abc customer",
			"description" : ""
			"shipping":{
				"address_line_1:"address line 1",
				"address_line_2:"can be empty",
				"admin_area_1:"can be empty",
				"admin_area_2:"can be empty",
				"postal_code:"200041",
				"country_code:"US"
			}
			}
		],

	payer: xxx 表示是否用那个支付平台，比如paypal
}
*/
func createPayment(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User create the paypal payment")
	//try to get user post information about the payment
	IP := GetIP(r)
	reqJSON := getJSONFromBody(r)

	payer := fmt.Sprint(reqJSON["payer"])
	if len(payer) == 0 || payer != "paypal" {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "参数错误：payer ",
		})
		return
	}

	items, ok := reqJSON["item_list"].([]interface{})
	if !ok {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "参数错误，没有设定item_list",
		})
		return
	}
	email := fmt.Sprint(reqJSON["email"])
	if len(email) == 0 {
		renderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "参数错误，没有设定email",
		})
		return
	}
	purchaseReqArray := make([]PurchaseUnitRequest, 0)
	// 循环处理item 内容
	for _, item := range items {
		citem, ok := item.(map[string]interface{})
		if !ok {
			renderJSON(w, r, map[string]interface{}{
				"retCode": -1,
				"msg":     "参数错误，item_list 结构错误",
			})
			return
		}
		invoiceID := getOrderID()
		customerID := fmt.Sprintf("%s", citem["customer_id"])

		description, ok := citem["description"].(string)
		if !ok {
			description = ""
		}
		refID, ok := citem["reference_id"].(string)
		if !ok {
			refID = ""
		}
		softDescriptor, ok := citem["soft_descriptor"].(string)
		if !ok {
			softDescriptor = ""
		}
		amountraw, ok := citem["amount"] //.(paypal.PurchaseUnitAmount)
		if !ok {
			renderJSON(w, r, map[string]interface{}{
				"retCode": -1,
				"msg":     "参数错误，没有设定amount",
			})
			return
		}
		amounttran, ok := amountraw.(map[string]interface{})

		currency, ok := amounttran["currency_code"].(string)
		if !ok {
			renderJSON(w, r, map[string]interface{}{
				"retCode": -1,
				"msg":     "参数错误，amount 参数错误",
			})
			return
		}
		value, ok := amounttran["value"].(string)
		if !ok {
			renderJSON(w, r, map[string]interface{}{
				"retCode": -1,
				"msg":     "参数错误，amount 参数错误",
			})
			return
		}

		amount := PurchaseUnitAmount{
			Currency: currency,
			Value:    value,
		}

		breakdown := amounttran["breakdown"] //.(paypal.PurchaseUnitAmountBreakdown)
		if !isNil(breakdown) {
			mm, err := json.Marshal(breakdown)
			if err != nil {
				logger.Warn("no breakdown ")
			} else {
				amount.Breakdown = &PurchaseUnitAmountBreakdown{}
				err = json.Unmarshal(mm, amount.Breakdown)
			}
		}
		//fmt.Println("========>", amount)

		sp, ok := citem["shipping"].(map[string]interface{})

		if !ok {
			renderJSON(w, r, map[string]interface{}{
				"retCode": -1,
				"msg":     "参数错误，没有设定shipping",
			})
			return
		}
		m, err := json.Marshal(sp)
		if err != nil {
			logger.Error(err)
			renderJSON(w, r, map[string]interface{}{
				"retCode": -1,
				"msg":     "参数错误，shipping 错误",
			})
			return
		}
		shipping := ShippingDetail{}
		err = json.Unmarshal(m, &shipping)
		if err != nil {
			logger.Error(err)
			renderJSON(w, r, map[string]interface{}{
				"retCode": -1,
				"msg":     "参数错误，shipping 错误",
			})
			return
		}

		onePurchaseReq := PurchaseUnitRequest{
			Amount: &amount,
		}

		if !isNil(invoiceID) {
			onePurchaseReq.InvoiceID = invoiceID
		}
		if !isNil(shipping) {
			onePurchaseReq.Shipping = &shipping
		}
		if !isNil(customerID) {
			onePurchaseReq.CustomID = customerID
		}
		if !isNil(description) {
			onePurchaseReq.Description = description
		}
		if !isNil(refID) {
			onePurchaseReq.ReferenceID = refID
		}
		if !isNil(softDescriptor) {
			onePurchaseReq.SoftDescriptor = softDescriptor
		}

		itemArray := make([]Item, 0)
		purchaseUnits := citem["items"].([]interface{})
		for _, ite := range purchaseUnits {

			//	m, ok := ite.(map[string]interface{})

			m, err := json.Marshal(ite)
			if err != nil {
				logger.Warn("item convert error")
				continue
			}
			realitem := Item{}
			ok := json.Unmarshal(m, &realitem)
			if ok == nil {
				itemArray = append(itemArray, realitem)
			} else {
				logger.Warn("item convert error")
				fmt.Println(ok)
				continue
			}

		}
		if len(itemArray) > 0 {
			onePurchaseReq.Items = itemArray
		}
		purchaseReqArray = append(purchaseReqArray, onePurchaseReq)
	}

	orderPayer := &CreateOrderPayer{
		EmailAddress: email,
	}

	//paymentList
	intent := OrderIntentCapture

	context := &ApplicationContext{
		ReturnURL: returnURL,
		CancelURL: cancelURL,
		BrandName: brand_name,
	}

	logger.Debug("Ready to create a order,Data Save to tempdb!")

	//	client.Lock()
	order, err := payClient.CreateOrder(intent, purchaseReqArray, orderPayer, context)
	//	client.Unlock()
	if err != nil {
		logger.Debug("create paypal order error", err)
		renderJSON(w, r, map[string]interface{}{
			"retCode": -2,
			"msg":     "创建订单失败",
		})
		return
	}
	//logger.Debug(order)

	if order.Status == PaypalCreated {
		go saveOrderRequest(*order, reqJSON, IP)
		lk := order.Links
		rd := ""
		for _, item := range lk {
			if item.Rel == "approve" {
				rd = item.Href
				break
			}
		}
		retData := map[string]interface{}{
			"transaction":  order,
			"redirect_url": rd,
		}
		renderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    retData,
		})
	} else {
		logger.Error(order)
		renderJSON(w, r, map[string]interface{}{
			"retCode": -10,
			"msg":     "Something error ",
			"data":    order.Status,
		})
	}

}

func excutePayment(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User excute the payment")
}

/*
respond to paypal redirection
https://view.bk.cloudns.cc:9090/return?paymentId=PAYID-L252FWA1FP659373E115460G&token=EC-3GB18191DL413423V&PayerID=47KPHAQQTU9A2
http://view.bk.cloudns.cc:9090/payment/paypal/return?token=6NE30544U39248243&PayerID=Y2BL7TKC2VUS4
to capture result is like
{"id":"4JM12086DK279151D","intent":"CAPTURE","purchase_units":[{"reference_id":"default","amount":{"currency_code":"USD","value":"30.11"},"payee":{"email_address":"sb-opduc1687278@business.example.com","merchant_id":"VRS5GBP9ETXXU"},"shipping":{"name":{"full_name":"subs one"},"address":{"address_line_1":"1 Main St","admin_area_2":"San Jose","admin_area_1":"CA","postal_code":"95131","country_code":"US"}},"payments":{"captures":[{"id":"4RD05868NB476373R","status":"COMPLETED","amount":{"currency_code":"USD","value":"30.11"},"final_capture":true,"seller_protection":{"status":"ELIGIBLE","dispute_categories":["ITEM_NOT_RECEIVED","UNAUTHORIZED_TRANSACTION"]},"seller_receivable_breakdown":{"gross_amount":{"currency_code":"USD","value":"30.11"},"paypal_fee":{"currency_code":"USD","value":"1.17"},"net_amount":{"currency_code":"USD","value":"28.94"}},"links":[{"href":"https://api.sandbox.paypal.com/v2/payments/captures/4RD05868NB476373R","rel":"self","method":"GET"},{"href":"https://api.sandbox.paypal.com/v2/payments/captures/4RD05868NB476373R/refund","rel":"refund","method":"POST"},{"href":"https://api.sandbox.paypal.com/v2/checkout/orders/4JM12086DK279151D","rel":"up","method":"GET"}],"create_time":"2020-05-14T04:58:25Z","update_time":"2020-05-14T04:58:25Z"}]}}],"payer":{"name":{"given_name":"subs","surname":"one"},"email_address":"subs1@usa.com","payer_id":"Y2BL7TKC2VUS4","address":{"country_code":"US"}},"create_time":"2020-05-14T04:57:23Z","update_time":"2020-05-14T04:58:25Z","links":[{"href":"https://api.sandbox.paypal.com/v2/checkout/orders/4JM12086DK279151D","rel":"self","method":"GET"}],"status":"COMPLETED"}

*/

func Succeed(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Paypal return from payment")
	q := r.URL.Query()
	payID := q.Get("PayerID")
	token := q.Get("token")
	logger.Debugf("User %s finished payment is down token is %s,", payID, token)
	tok := PaymentSourceToken{
		ID:   token,
		Type: "BILLING_AGREEMENT",
	}

	captureRes, err := payClient.CaptureOrder(token, CaptureOrderRequest{PaymentSource: &PaymentSource{
		Token: &tok,
	}})
	if err != nil {
		fmt.Println(err)
	}
	if captureRes.Status == PaypalCompleted {
		logger.Debug(captureRes)
		go saveCapture(*captureRes)
		// now to forward to succssful url
		w.Write([]byte(fmt.Sprintf("订单号 ：%s,Thanks！", captureRes.ID)))
	} else {
		logger.Debug("Something error")
		w.Write([]byte(fmt.Sprintf("可能有点问题，回到购物车！")))
	}

}

func Failed(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User Cancel the payment")

}

func Test(w http.ResponseWriter, r *http.Request) {
	ree := getJSONFromBody(r)
	//id, _ := ree["id"].(string)
	//data, _ := ree["status"].(string)
	//UpdateOrderStatus(id, data)
	logger.Debug("Get the input is:", ree)
}

func Index(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Index open!")
	w.Write([]byte("PayPal Index"))
}

// Notify accept all the notification from paypal
func Notify(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Notify from Paypal")
	bodybytes := getBinaryDataFromBody(r)
	logger.Debug(len(bodybytes))
	/* verifyResp, _ := payClient.VerifyWebhookSignature(r, webHookID)

	if verifyResp.VerificationStatus != PaypalVerified || ProgramMode != "DEBUG" {
		logger.Errorf("failed to decode request body: %s", verifyResp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Debugf("Verification pass") */
	ree := getJSONFromBody(r)

	byarry, _ := json.Marshal(ree)
	//logger.Debug("body:", string(bodybytes))
	//logger.Debug("map string ", string(byarry))
	var notification WebHookNotifiedEvent
	err := json.Unmarshal(byarry, &notification)
	if err != nil {
		logger.Errorf("failed to decode request body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//logger.Debug(PrettyPrint(notification))
	go saveNotify(notification) // Save the notification

	logger.Debugf("paypal notification %s arrival with resource %v", notification.ID, notification.EventType)
	switch {
	case notification.EventType == EventPaymentCaptureCompleted:

		//verifyResp, _ := payClient.VerifyWebhookSignatureWithData(r, webHookID, bodybytes)
		//logger.Debug(PrettyPrint(verifyResp))
		//if verifyResp.VerificationStatus == PaypalVerified || ProgramMode == "DEBUG" {
		if ProgramMode == "DEBUG" {
			resource := CaptureResource{}
			rr, _ := json.Marshal(notification.Resource)
			err := json.Unmarshal(rr, &resource)
			if err != nil {
				logger.Debug("Capture completed resource unmarshal failed", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			oid, ok := CreateNewOrderInDB(&notification, resource)
			if ok {
				logger.Info("New order Created!")
				//go data.UpdateOrderByID(fmt.Sprint("%d", oid), data.OrderInValidate)
				UpdateOrderStatusByOrderID(fmt.Sprint("%d", oid), data.OrderInValidate)
			} else {
				logger.Info("New order Created Error!")
			}
		} else {
			logger.Warn("The verification failed!")
		}
		w.WriteHeader(http.StatusOK)
	case notification.EventType == EventPaymentCaptureDenied:
		logger.Debug("Payment is denied: ", notification.ID)

		w.WriteHeader(http.StatusOK)
	case notification.EventType == EventPaymentCaptureRefunded:
		logger.Debug("Payment is refund: ", notification.ID)

		resource := RefundResource{}
		rr, _ := json.Marshal(notification.Resource)
		err := json.Unmarshal(rr, &resource)
		if err != nil {
			logger.Debug("refund resource unmarshal failed", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Debugf("refunded for payment id is %v", resource.ID)
		UpdateOrderStatusByPaymentID(resource.ID, data.OrderRefunded)
		w.WriteHeader(http.StatusOK)

	case notification.EventType == EventCustomerDisputeResolved:
		logger.Debugf("Customer dispute! : ", notification.ID)
		resource := DisputeResource{}
		rr, _ := json.Marshal(notification.Resource)
		err := json.Unmarshal(rr, &resource)
		if err != nil {
			logger.Debug("Dispute resource unmarshal failed", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Debugf("Disputed id is %s, sell transactions id is %v", resource.DisputeID, resource.DisputedTransactions)
		for _, item := range resource.DisputedTransactions {
			UpdateOrderStatusByPaymentID(item.SellTransactionID, data.OrderDisputed)
		}

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusOK)
	}

}
