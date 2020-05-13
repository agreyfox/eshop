package paypal

import (
	"github.com/agreyfox/eshop/admin"
	"github.com/plutov/paypal"
	"net/http"
	"os"
)

var (
	// paypal ClientID
	ClientID = "AbOMcM4iaf0PYKGgOFCktDD-Rqzpn7R_r2yPfwbopgCLYkBLXkD45c1qejwVX2BrBSxVQgz3_QlU7iFn"
	// Paypal client secrte
	Secret      = "EKxToL0apcJ7HOAryLeFkyP9JRWuw-p8pMj9M5N3Y1Ee8tsUDFgRv1wA_3hIjRMiHqrmbQu12KW_Noys"
	accessToken *paypal.TokenResponse
	returnURL   = "http://view.bk.cloudns.cc/payment/paypal/return"
	cancelURL   = "http://view.bk.cloudns.cc/payment/paypal/cancle"

	paypalClient *paypal.Client
)

func init() {
	paypalClient, err := paypal.NewClient(ClientID, Secret, paypal.APIBaseSandBox)
	paypalClient.SetLog(os.Stdout) // Set log to terminal stdout
	logger.Debug("Paypal get access token result:", err)
	accessToken, err = paypalClient.GetAccessToken()
	logger.Debug(accessToken)
}

// When user to checkout "Pay Now" button ,It will send the request to beckend system and beckend system will
// send the request to create the payment.return the created payment information with authorization url
/* input data is looks like
{

}
*/
func createPayment(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User create the payment")
	//try to get user post information about the payment
	reqJSON := admin.getJsonFromBody()
	logger.Debug(reqJSON)

}

func excutePayment(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User excute the payment")
}

func Succeed(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Paypal return from payment")
}

func Failed(w http.ResponseWriter, r *http.Request) {
	logger.Debug("User Cancle the payment")

}

// to save a record
func save(db *bolt.DB, payment interface{}) error {
	// Store the user model in the user bucket using the username as the key.
	err := dbHandler.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(usersBucket)

		if err != nil {
			return err
		}

		encoded, err := json.Marshal(user)
		if err != nil {
			return err
		}
		return b.Put([]byte(user.Name), encoded)
	})
	return err
}
