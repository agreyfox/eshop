package paypal

import (
	"encoding/json"
	"fmt"

	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/system/db"
	"github.com/go-zoo/bone"

	"net/http"
	"time"

	"github.com/robfig/cron"
)

const (
	prefer string = "return=representation"
)

var (
	brand_name string = "EGpal Game Center"
	// paypal ClientID
	ClientID = "AbOMcM4iaf0PYKGgOFCktDD-Rqzpn7R_r2yPfwbopgCLYkBLXkD45c1qejwVX2BrBSxVQgz3_QlU7iFn"
	//ClientID = "AQETpIU9UXmAro-k0aN6EZHHG-iGvbVdXQH4ywhrOmd8UiVWZF6YBvUkzR2MVJuoxFr2T6Q7kXTUVpEQ"
	//access_token$sandbox$7kgb9j9nsb5bfmbx$5fe304f1bbd2c4fad7c74438737f8cec
	// Paypal client secrte
	Secret = "EKxToL0apcJ7HOAryLeFkyP9JRWuw-p8pMj9M5N3Y1Ee8tsUDFgRv1wA_3hIjRMiHqrmbQu12KW_Noys"
	//Secret      = "EHrup4KHaN2QCa59_oIdaH4cl2u5fYMW5b-g5h99T9YOsyq4mRldit8cztRJ-N1xzMcQ_oBoCdTOSp33"
	accessToken *TokenResponse
	returnURL   = "https://support.bk.cloudns.cc:8081/payment/paypal/return"
	cancelURL   = "https://support.bk.cloudns.cc:8081/payment/paypal/cancel"
	thanksURL   = "https://support.bk.cloudns.cc:8081/thanks"
	webHookID   = "3S402195M8327334K"
	//webHookHD   = "30G82858PK903472T"
	hookURL     = "https://support.bk.cloudns.cc:8081/payment/paypal/notify"
	payClient   *Client
	ProgramMode = "DEBUG"
)

func initpaypal() {

	key, err := db.GetParameterFromConfig("PaymentSetting", "name", "company_name", "valueString")
	if err == nil {
		brand_name = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "paypal_ClientID", "valueString")
	if err == nil {
		ClientID = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "paypal_Secret", "valueString")
	if err == nil {
		Secret = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "paypal_webHookID", "valueString")
	if err == nil {
		webHookID = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "paypal_NotifyURL", "valueString")
	if err == nil {
		hookURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "paypal_returnURL", "valueString")
	if err == nil {
		returnURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "paypal_cancelURL", "valueString")
	if err == nil {
		cancelURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "paypal_ProgramMode", "valueString")
	if err == nil {
		ProgramMode = key
	}
	apibase := APIBaseSandBox
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "paypal_ApiBase", "valueString")
	if err == nil {
		apibase = key
	}

	client, err := NewClient(ClientID, Secret, apibase)
	client.SetLog(logger) // Set log to terminal stdout
	payClient = client

	logger.Debug("Paypal get access token result:", err)
	accessToken, err = payClient.GetAccessToken()
	//logger.Debug(accessToken)

	logger.Info("Start Paypal access token refresh job ")
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

func userSubmit(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User submit a  paypal  payment")
	//ip := GetIP(r)
	//try to get user post information about the payment

	payload := new(data.UserSubmitOrderRequest)
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {

		logger.Errorf("user submit error", err)
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     err,
		})
		return
	}
	//reqJSON := getJSONFromBody(r)
	payload.IPAddr = data.GetIP(r)
	if err := validateRequest(payload); err != nil {
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     err,
		})
		return
	}
	payload.OrderID = data.GetShortOrderID()
	payload.OrderDate = time.Now().Unix()
	respond, err := createOrder(payload) //create payssion call

	rettxt, _ := json.MarshalIndent(respond, "", "  ")
	payload.Respond = string(rettxt)

	errcreateorder := data.SaveOrderRequest(payload) //finished save request,
	logger.Infof("Create paypal request in db with err:", errcreateorder)
	if err != nil {
		logger.Error("create paypal order error:", err)
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -2,
			"msg":     err,
		})
		return
	}

	if respond.Status == PaypalCreated {
		logger.Info("Pay Payment Step 1 done")
		//go saveOrderRequest(*order, orderReq, IP)
		lk := respond.Links
		rd := ""
		for _, item := range lk { //get the approve link for customer approve
			if item.Rel == "approve" {
				rd = item.Href
				break
			}
		}
		retData := map[string]interface{}{
			"transaction":  respond,
			"redirect_url": rd,
		}
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    retData,
		})
	} else {
		logger.Error(respond)
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -10,
			"msg":     "Something error,Please retry later",
			"data":    respond.Status,
		})
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
	IP := data.GetIP(r)
	//reqJSON := getJSONFromBody(r)
	rawData := data.GetBinaryDataFromBody(r) //support.bk.cloudns.cc:8081/paypment/static/return

	orderReq, err := getOrderRequestFromUserSubmit(rawData)
	if err != nil {
		logger.Error(err)
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     err.Error(),
		})
		return
	}
	//fmt.Println(or, err)
	//orderReq := *or
	invoiceID := data.GetShortOrderID()

	err = json.Unmarshal(rawData, &orderReq)
	if err != nil {
		logger.Error(err)
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -1,
			"msg":     "参数错误:" + err.Error(),
		})
		return
	}

	for i := range orderReq.ItemList {
		orderReq.ItemList[i].InvoiceID = invoiceID
		orderReq.ItemList[i].Description = GeneratePurchaseContentFromRequest(&orderReq.ItemList[i])
	}

	orderPayer := &CreateOrderPayer{
		EmailAddress: orderReq.Email,
	}

	//paymentList
	intent := OrderIntentCapture
	lo := orderReq.Locale
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
	if orderReq.Method == LandingPageTypeBilling {
		context.LandingPage = LandingPageTypeBilling // support credut card payment.
	}
	logger.Debug("Ready to create a order,Data Save to tempdb!")

	//	client.Lock()
	//logger.Debugf("This item_list is %v", orderReq.ItemList)

	order, err := payClient.CreateOrder(intent, orderReq.ItemList, orderPayer, context)
	//	client.Unlock()
	if err != nil {
		logger.Debug("create paypal order error:", err)
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -2,
			"msg":     "创建订单失败",
		})
		return
	}
	//logger.Debug(order)

	if order.Status == PaypalCreated {
		logger.Info("Payment Step 1 done")
		go saveOrderRequest(*order, orderReq, IP)
		lk := order.Links
		rd := ""
		for _, item := range lk { //get the approve link for customer approve
			if item.Rel == "approve" {
				rd = item.Href
				break
			}
		}
		retData := map[string]interface{}{
			"transaction":  order,
			"redirect_url": rd,
		}
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    retData,
		})
	} else {
		logger.Error(order)
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -10,
			"msg":     "Something error,Please retry later  ",
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

	q := r.URL.Query()
	payID := q.Get("PayerID")
	token := q.Get("token")
	logger.Infof("Paypal return from order creation,Step 2 done\n User %s finished payment , the token is %s,", payID, token)
	tok := PaymentSourceToken{
		ID:   token,
		Type: "BILLING_AGREEMENT",
	}
	captureRes, err := payClient.CaptureOrder(token, CaptureOrderRequest{PaymentSource: &PaymentSource{
		Token: &tok,
	}})
	if err != nil {
		logger.Error("Capture Order has some error ", err)
		url := data.OnlineURL
		url += fmt.Sprintf("status=0&msg=%s", err.Error())
		http.Redirect(w, r, url, http.StatusFound)
	}
	if captureRes.Status == PaypalCompleted {
		logger.Info("Paypal payment Step 3 to send capture payment is done!", captureRes.ID)
		logger.Debug("The Capture Result:", captureRes)
		go saveCapture(*captureRes) //to save capture

		url := data.OnlineURL
		if len(captureRes.ID) > 0 {
			url += fmt.Sprintf("?status=1&orderno=%s", captureRes.PurchaseUnits[0].InvoiceID)
		} else {
			url += fmt.Sprintf("status=0&msg=%s", "Payment finished with error")
		}
		http.Redirect(w, r, url, http.StatusFound)
	} else {
		logger.Debug("Paypal Step 3 send capture order   return with error:", captureRes.Status)
		url := data.OnlineURL
		url += fmt.Sprintf("status=0&msg=%s", captureRes.Status)
		http.Redirect(w, r, url, http.StatusFound)
	}

}

func Failed(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User Cancel the payment")

}

func Test(w http.ResponseWriter, r *http.Request) {
	ree := data.GetJSONFromBody(r)
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

	bodybytes := data.GetBinaryDataFromBody(r)
	logger.Debug("Notification from Paypal:", string(bodybytes[:]))

	ree := data.GetJSONFromBody(r)

	byarry, err := json.Marshal(ree)
	if err != nil {
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": -88,
			"msg":     "error when reterive notify data",
		})
		return
	}
	var notification WebHookNotifiedEvent
	err = json.Unmarshal(byarry, &notification)
	if err != nil {
		logger.Errorf("failed to decode notify body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go saveNotify(notification) // Save the notification

	logger.Infof("paypal notification %s arrival with resource %v", notification.ID, notification.EventType)
	switch {
	case notification.EventType == EventPaymentCapturePending:
		logger.Infof("Step 4 Capture after customer approve,The resource is :%s", fmt.Sprint(notification.Resource))
		//verifyResp, _ := payClient.VerifyWebhookSignatureWithData(r, webHookID, bodybytes)
		//logger.Debug(verifyResp)
		//	ProgramMode = "DEBUG"
		//if verifyResp.VerificationStatus == PaypalVerified || ProgramMode == "DEBUG" {
		if ProgramMode == "DEBUG" {
			resource := Resource{}
			rr, _ := json.Marshal(notification.Resource)
			err := json.Unmarshal(rr, &resource)
			if err != nil {
				logger.Debug("Notification checkout resource unmarshal failed", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			//logger.Debug("capture resource is ", resource)
			if len(resource.Amount.Value) == 0 || len(resource.InvoiceID) == 0 {
				logger.Debug("The return checkout resource has no valid data ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			oid, _, ok := CreateNewPendingOrderInDB(&notification, resource)

			if ok {
				logger.Infof("New paypal Pending order %d Created!", oid)

			} else {
				logger.Info("New order Created Error!")
			}

		} else {
			logger.Warn("The paypal webhook  verification failed!")
		}
		w.WriteHeader(http.StatusOK)
	case notification.EventType == EventCheckOrderApproved:
		logger.Infof("Step 4 Capture after customer approve,The resource is :%s", fmt.Sprint(notification.Resource))
		//verifyResp, _ := payClient.VerifyWebhookSignatureWithData(r, webHookID, bodybytes)
		//logger.Debug(verifyResp)
		if ProgramMode == "DEBUG" {
			//if verifyResp.VerificationStatus == PaypalVerified || ProgramMode == "DEBUG" {
			//if ProgramMode == "DEBUG" {
			//resource := CaptureResource{}
			resource := Resource{}
			rr, _ := json.Marshal(notification.Resource)
			err := json.Unmarshal(rr, &resource)
			if err != nil {
				logger.Debug("Notification checkout resource unmarshal failed", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			//logger.Debug("capture resource is ", resource)
			if len(resource.PurchaseUnits) == 0 {
				logger.Debug("The return checkout resource has no valid data ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			oid, oo, ok := CreateNewOrderInDB(&notification, resource)
			if ok {
				if len(oo.Payer) > 0 {
					orderid := oo.OrderID
					logger.Debugf("paypal send eamil for order accept %s,%s", oo.Payer, GetPurchaseContent(orderid))
					go data.SendConfirmEmail(orderid, GetPurchaseContent(orderid), resource.PurchaseUnits[0].Amount.Value, resource.PurchaseUnits[0].Amount.Currency, oo.Payer)
				}
				logger.Infof("New paypal order %d Created!", oid)

			} else {
				logger.Info("New order Created Error!")
			}
			//go SaveTransationDetail(resource.ID)
		} else {
			logger.Warn("The paypal webhook  verification failed!")
		}
		w.WriteHeader(http.StatusOK)

	case notification.EventType == EventPaymentCaptureCompleted:
		logger.Debug("The new payment order capture approved", notification.ID)

	case notification.EventType == EventPaymentOrderCreated:
		logger.Debug("The new payment order created", notification.ID)

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

/// TranscationInfo reterive transactions infomation to admin user
func TransactionInfo(w http.ResponseWriter, r *http.Request) {
	tid := bone.GetValue(r, "id")
	logger.Infof("Search transactions id :%s", tid)

	result, ok := GetTransationDetail(tid)
	if ok == nil {
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    result,
		})
	} else {
		logger.Error("Search error,Request from :", data.GetIP(r))
		data.RenderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "internal error:" + ok.Error(),
			"data":    result,
		})
	}
}
