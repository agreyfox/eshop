package user

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/agreyfox/eshop/system/admin/user"
	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/logs"

	//jwt "github.com/dgrijalva/jwt-go"
	emailer "github.com/nilslice/email"

	"github.com/nilslice/jwt"
	"go.uber.org/zap"
)

var (
	err    error
	logger *zap.SugaredLogger = logs.Log.Sugar()
)

func RegisterUsersHandler(res http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case http.MethodPost:
		// create new user
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB

		if err != nil {
			log.Println(err)
			renderJSON(res, req, RetUser{RetCode: -1, Msg: err.Error()})
			return
		}

		email := strings.ToLower(req.FormValue("email"))
		password := req.PostFormValue("password")

		if email == "" || password == "" {

			renderJSON(res, req, RetUser{
				RetCode: -21,
				Msg:     "Wrong Data"})
			return
		}

		usr, err := user.NewCustomer(email, password)
		if err != nil {
			log.Println(err)
			renderJSON(res, req,
				RetUser{
					RetCode: -1,
					Msg:     err.Error(),
					Data:    "",
				})
			return
		}

		_, err = db.SetUser(usr)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		//http.Redirect(res, req, req.URL.String(), http.StatusFound)
		//res.WriteHeader(http.StatusAccepted)

		renderJSON(res, req, RetUser{
			RetCode: 0,
			Msg:     "Done",
		})

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getJsonFromBody(req *http.Request) map[string]interface{} {
	var body string
	bodyBytes, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		body = fmt.Sprintf("failed read Read response body: %v", err)
		logger.Debug(body)
		return map[string]interface{}{}
	}
	var t map[string]interface{}

	//fmt.Println(body)

	if err = json.Unmarshal(bodyBytes, &t); err != nil {

		return map[string]interface{}{}
	}
	return t

}

func LoginHandler(res http.ResponseWriter, req *http.Request) {
	//logger.Debugf("%v", req)
	switch req.Method {

	case http.MethodPost:
		if user.IsValid(req) {
			logger.Debug("is valid")
			//http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			renderJSON(res, req,
				RetUser{
					RetCode: 2,
					Msg:     "Already Loggin",
					Data:    "",
				})
			return
		}
		requestJson := getJsonFromBody(req)
		//fmt.Println(requestJson)
		/* err := req.ParseForm()

		if err != nil {
			log.Println(err)
			renderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     err.Error(),
				Data:    "",
			})
			return
		}
		*/
		// check email & password
		email := requestJson["email"].(string)
		password := requestJson["password"].(string)
		logger.Debug("The Request email is :", email)
		j, err := db.User(strings.ToLower(email))

		if err != nil {
			log.Println(err)
			renderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     err.Error(),
				Data:    "",
			})
			return
		}

		if j == nil {
			renderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     "no such user"})
			return
		}

		usr := &user.User{}
		err = json.Unmarshal(j, usr)
		if err != nil {
			log.Println(err)
			renderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     err.Error(),
				Data:    "",
			})
			return
		}

		if !user.IsUser(usr, password) {
			renderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     "Wrong email or password!",
				Data:    "",
			})
			return
		}
		// create new token
		week := time.Now().Add(time.Hour * 2) // session time is 2 hours

		claims := map[string]interface{}{
			"exp":  week,
			"user": usr.Email,
		}
		token, err := jwt.New(claims)
		//DecodeJwt(token)
		logger.Debug(jwt.GetClaims(token))

		if err != nil {
			log.Println(err)
			renderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     "Internal Error",
				Data:    "",
			})
			return
		}

		// add it to cookie +1 week expiration
		http.SetCookie(res, &http.Cookie{
			Name:    user.Lqcmstoken,
			Value:   token,
			Expires: week,
			Path:    "/",
		})

		logger.Debugf("User %s logged in !", usr)
		renderJSON(res, req, RetUser{
			RetCode: 0,
			Msg:     "Done",
			Data:    token,
		})

		return
		//http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/login"), http.StatusFound)
	}
}

func RenewHandler(res http.ResponseWriter, req *http.Request) {
	logger.Debugf("%v", req)
	switch req.Method {

	case http.MethodGet:
		if user.IsValid(req) {
			logger.Debug("is valid")
			//http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			week := time.Now().Add(time.Hour * 2) // session time is 2 hours
			// check if token exists in cookie
			cookie, err := req.Cookie(user.Lqcmstoken)
			if err == nil {

				// validate it and allow or redirect request
				token := cookie.Value
				http.SetCookie(res, &http.Cookie{
					Name:    user.Lqcmstoken,
					Value:   token,
					Expires: week,
					Path:    "/",
				})
				logger.Debug("User renew !")
				renderJSON(res, req, RetUser{
					RetCode: 0,
					Msg:     "Done",
					Data:    "",
				})
			} else {
				renderJSON(res, req, RetUser{
					RetCode: -10,
					Msg:     err.Error(),
				})
			}
		} else {
			renderJSON(res, req, RetUser{
				RetCode: -9,
				Msg:     "not valid user",
			})
		}
		return
	default:
		//		logger.Debugf("User %s logged in !", usr)
		renderJSON(res, req, RetUser{
			RetCode: -10,
			Msg:     "failed",
			Data:    "",
		})

		return
	}
}

func LogoutHandler(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:    user.Lqcmstoken,
		Expires: time.Unix(0, 0),
		Value:   "",
		Path:    "/",
	})
	renderJSON(res, req, RetUser{
		RetCode: 0,
		Msg:     "Done",
		Data:    "",
	})
	return
	//	http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/login", http.StatusFound)
}

func ForgotPasswordHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			renderJSON(res, req, RetUser{
				RetCode: -2,
				Msg:     err.Error(),
				Data:    "",
			})
			return
		}

		// check email for user, if no user return Error
		email := strings.ToLower(req.FormValue("email"))
		if email == "" {
			res.WriteHeader(http.StatusBadRequest)
			log.Println("Failed account recovery. No email address submitted.")
			return
		}

		_, err = db.User(email)
		if err == db.ErrNoUserExists {
			res.WriteHeader(http.StatusBadRequest)
			log.Println("No user exists.", err)
			return
		}

		if err != db.ErrNoUserExists && err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Println("Error:", err)
			return
		}

		// create temporary key to verify user
		key, err := db.SetRecoveryKey(email)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed to set account recovery key.", err)
			return
		}

		domain, err := db.Config("domain")
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed to get domain from configuration.", err)
			return
		}

		body := fmt.Sprintf(`
There has been an account recovery request made for the user with email:
%s

To recover your account, please go to http://%s/admin/recover/key and enter 
this email address along with the following secret key:

%s

If you did not make the request, ignore this message and your password 
will remain as-is.


Thank you,
%s

`, email, domain, key, domain)

		msg := emailer.Message{
			To:      email,
			From:    fmt.Sprintf("admin@%s", domain),
			Subject: fmt.Sprintf("Account Recovery [%s]", domain),
			Body:    body,
		}

		go func() {
			err = msg.Send()
			if err != nil {
				log.Println("Failed to send message to:", msg.To, "about", msg.Subject, "Error:", err)
			}
		}()

		// redirect to /admin/recover/key and send email with key and URL
		//http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/recover/key", http.StatusFound)
		renderJSON(res, req, RetUser{RetCode: 0, Msg: "Done"})
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func RecoveryKeyHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println("Error parsing recovery key form:", err)

			renderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again."})
			return
		}

		// check for email & key match
		email := strings.ToLower(req.FormValue("email"))
		key := req.FormValue("key")

		var actual string
		if actual, err = db.RecoveryKey(email); err != nil || actual == "" {
			log.Println("Error getting recovery key from database:", err)
			renderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again."})
			return
		}

		if key != actual {
			log.Println("Bad recovery key submitted:", key)

			renderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
			return
		}

		// set user with new password
		password := req.FormValue("password")
		usr := &user.User{}
		u, err := db.User(email)
		if err != nil {
			log.Println("Error finding user by email:", email, err)

			renderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
			return
		}

		if u == nil {
			log.Println("No user found with email:", email)

			renderJSON(res, req, RetUser{RetCode: -1, Msg: "Error,  please go back and try again.", Data: ""})
			return
		}

		err = json.Unmarshal(u, usr)
		if err != nil {
			log.Println("Error decoding user from database:", err)

			renderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
			return
		}

		update, err := user.NewCustomer(email, password)
		if err != nil {
			log.Println(err)

			renderJSON(res, req, RetUser{RetCode: -1, Msg: err.Error(), Data: ""})
			return
		}

		update.ID = usr.ID

		err = db.UpdateUser(usr, update)
		if err != nil {
			log.Println("Error updating user:", err)
			renderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
			return
		}

		renderJSON(res, req, RetUser{RetCode: 1, Msg: "Done,Pleaes relogin", Data: ""})

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// Auth is HTTP middleware to ensure the request has proper token credentials
func CustomerAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		if user.IsValid(req) {
			next.ServeHTTP(res, req)
		} else {
			res.WriteHeader(http.StatusForbidden)
			logger.Error("Action %s without user permission:", req.RequestURI)
			return
		}
	})
}

/*
func generateToken(usr string, week time.Time) (string, error) {
	var claims map[string]interface{}{
		"user":usr,
		"exp":week,
		"StandardClaims": jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 2).Unix(),
			Issuer:    "EShop",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(d.settings.Key)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "cty")
	w.Write([]byte(signed))
	return 0, nil
}
*/
