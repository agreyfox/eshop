package static

import (
	"encoding/json"
	"html"
	"time"

	"net/http"

	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/system/db"
)

var (
	payClient *Client
	payMethod string
)

func initStatic() {

	htpage, err := db.GetParameterFromConfig("PaymentSetting", "name", "static_url", "valueString")
	if err == nil {
		StaticURL = htpage
	}
	htpage, err = db.GetParameterFromConfig("PaymentSetting", "name", "static_pay", "valueString")
	if err == nil {
		forwardpage = htpage
	}
	htpage, err = db.GetParameterFromConfig("PaymentSetting", "name", "static_url", "description")
	if err == nil {

		paymentpage = html.UnescapeString(htpage)
	} else {
		logger.Warn("paymetpage is not setting,Please check!")
	}

	payClient = &Client{}

}

// accept user standard request
func userSubmit(w http.ResponseWriter, r *http.Request) {
	logger.Info("User submit a  static  payment")

	payload := new(data.UserSubmitOrderRequest)
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
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
	logger.Debugf("Create static payment order, error:&s", err)
	payload.Respond = respond

	//errsave := data.SaveOrderRequest(payload) //finished save request,

	//	logger.Debug(errsave)
	/* var proto = ""
	if strings.HasPrefix(r.Proto, "HTTPS") {
		proto = "https://"
	} else {
		proto = "http://"
	} */
	retData := map[string]interface{}{
		"transaction":  payload,
		"redirect_url": StaticURL + payClient.GetPaymentPage(),
	}
	data.RenderJSON(w, r, map[string]interface{}{
		"retCode": 0,
		"msg":     "ok",
		"data":    retData,
	})

}

func Succeed(w http.ResponseWriter, r *http.Request) {

	logger.Debugf("Call to static payment page /payment/static/paypage")

	http.Redirect(w, r, StaticURL+"/payment/static/paypage", http.StatusFound)
}

func Notify(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Get Notify data from Static page:", data.GetIP(r))
	bodybytes := data.GetBinaryDataFromBody(r)
	logger.Debug(string(bodybytes[:]))

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
