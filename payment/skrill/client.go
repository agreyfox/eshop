package skrill

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/agreyfox/eshop/system/db"
)

/*
 1 hen the customer is ready to pay for goods or services on your website, they select the Skrill payment option on your website.
2. You request a session identifier (SID) by passing customer and transaction details (for example: amount, currency and language) to Quick Checkout.
3. Skrill returns the generated SID.
4. Using a light box or iframe you redirect the customer to Quick Checkout and include the session identifier in the redirect URL. Skrill displays the Quick Checkout page.
5. The customer enters their payment information, plus any other details requested, and confirms the transaction.
6. Skrill requests authorisation for the payment from the customer’s bank, third party provider or card issuer.
7. The bank/provider approves or rejects the transaction.
8. Skrill displays the Quick Checkout confirmation page, containing the transaction result.
9. Skrill provides you with an asynchronous notification, sent to your status URL or IPN (instant Payment Notification), confirming the transaction details and status
*/
var (
	defaultURL     = "https://pay.skrill.com/"
	payURL         = "https://www.skrill.com/app/pay.pl"
	PrepareAction  = "action=prepare"
	TransferAction = "action=transfer"
	MerchantEmail  = "e_raeb@163.com"
	Password       = "tqxy0605"
	MerchantID     = "138853317"
	Subject        = "Egpal game shop"
	NotifyURL      = "https://support.bk.cloudns.cc:8081/payment/skrill/notify"
	ReturnURL      = "https://support.bk.cloudns.cc:8081/payment/skrill/return"
	CancelURL      = "https://support.bk.cloudns.cc:8081/payment/skrill/cancel"
)

func initSkrill() {

	key, err := db.GetParameterFromConfig("PaymentSetting", "name", "skrill_merchantEmail", "valueString")
	if err == nil {
		MerchantEmail = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "skrill_password", "valueString")
	if err == nil {
		Password = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "skrill_merchantID", "valueString")
	if err == nil {
		MerchantID = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "skrill_subject", "valueString")
	if err == nil {
		Subject = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "skrill_NotifyURL", "valueString")
	if err == nil {
		NotifyURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "skrill_ReturnURL", "valueString")
	if err == nil {
		ReturnURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "skrill_CancelURL", "valueString")
	if err == nil {
		CancelURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "skrill_Server", "valueString")
	if err == nil {
		defaultURL = key
	}
	payClient = New() //create skrill instance
	logger.Infof("Skrill backend service initialized! MerchantID is %s", MerchantID)
}

// Client is a Skrill client
type Client struct {
	url        string
	email      string
	merchantid string
	secretWord string
}

// New initiates Skrill client
func New(configs ...Config) *Client {
	client := &Client{
		url:        defaultURL,
		email:      MerchantEmail,
		secretWord: Password,
		merchantid: MerchantID,
	}

	if len(configs) != 0 {
		config := configs[0]
		client.url = config.URL
		client.email = config.Email
		client.secretWord = config.SecretWord
		client.merchantid = config.MerchantID
	}
	return client
}

// Prepare make a request to prepare payment and returns redirect url
func (c *Client) Prepare(param *PrepareParam) (redirectURL string, err error) {
	param.PrepareOnly = "1"
	if len(param.PayToEmail) == 0 {
		param.PayToEmail = c.email
	}
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(param)
	res, err := http.Post(c.url, "application/json", b)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	//logger.Debug(res)
	if res.StatusCode != http.StatusOK {
		var err ErrSkrill
		if e := json.NewDecoder(res.Body).Decode(&err); e != nil {
			return "", e
		}
		return "", err
	}

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	//logger.Debug(string(bs[:]))
	return genRedirectURL(c.url, string(bs)), nil
}

func genRedirectURL(url, sessionID string) string {
	return url + "?sid=" + sessionID
}
