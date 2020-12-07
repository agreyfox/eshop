package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/agreyfox/eshop/system/admin/user"
	smtp2go "github.com/agreyfox/eshop/system/email"
	"github.com/agreyfox/eshop/system/ip"

	"github.com/agreyfox/eshop/system/db"

	"github.com/nilslice/jwt"
)

func RegisterUsersHandler(res http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case http.MethodPost:
		// create new user
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB

		if err != nil {
			log.Println(err)

			RenderJSON(res, req, RetUser{RetCode: -1, Msg: err.Error()})
			return
		}

		email := strings.ToLower(req.FormValue("email"))
		password := req.PostFormValue("password")

		if email == "" || password == "" {

			RenderJSON(res, req, RetUser{
				RetCode: -21,
				Msg:     "Wrong Data"})
			return
		}

		usr, err := user.NewCustomer(email, password)
		if err != nil {
			log.Println(err)
			RenderJSON(res, req,
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

		RenderJSON(res, req, RetUser{
			RetCode: 0,
			Msg:     "Done",
		})

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// for user register , and return new user id
func RegisterUser(res http.ResponseWriter, req *http.Request) {
	logger.Debugf("User try to register , from:", GetIP(req))
	reqJSON := GetJsonFromBody(req)

	email := strings.ToLower(reqJSON["email"].(string))
	password := reqJSON["password"].(string)
	social := ""
	//phone := ""
	meta := ""
	social = reqJSON["social_link"].(string)
	//phone = reqJSON["phone"].(string)
	meta = reqJSON["social_type"].(string)

	if email == "" || password == "" {
		logger.Error("Insufficient user register information")
		RenderJSON(res, req, RetUser{
			RetCode: -21,
			Msg:     "Wrong Register User Data"})
		return
	}

	usr, err := user.NewCustomerWithSocial(email, password, meta, social)
	if err != nil {
		logger.Error(err)
		RenderJSON(res, req,
			RetUser{
				RetCode: -1,
				Msg:     err.Error(),
				Data:    "",
			})
		return
	}
	//usr.Phone = phone
	usr.Social = social
	usr.Meta = meta
	_, err = db.SetUser(usr)
	if err != nil {
		logger.Error(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	//http.Redirect(res, req, req.URL.String(), http.StatusFound)
	//res.WriteHeader(http.StatusAccepted)

	RenderJSON(res, req, RetUser{
		RetCode: 0,
		Msg:     "Done",
		Data:    usr.ID, // return user id
	})

}

// for user register , and return new user id
/*
update for user profile,
	request like:
	{
		"email":"abc@mail.com",
		"newPassword:"sadfas",
		"social_link":"ssdmaasfd",
		"social_type":"phone"
	}
*/
func UpdateUser(res http.ResponseWriter, req *http.Request) {
	ipAddr := GetIP(req)
	logger.Debugf("User try to update user account , from:", ipAddr)
	//reqJSON := GetJsonFromBody(req)
	if !user.IsValid(req) {
		logger.Error("invalid user update profile")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	userReq := new(UpdateUserRequest)
	err := json.NewDecoder(req.Body).Decode(userReq)
	if err != nil {
		logger.Error("invalid user request data")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	email := strings.ToLower(userReq.Email)
	password := strings.Trim(userReq.NewPassword, " ,")
	social := ""
	meta := ""
	social = strings.Trim(userReq.Social, " ,")
	meta = strings.Trim(userReq.Type, " ,")

	if email == "" || (password == "" && (len(social) == 0 || len(meta) == 0)) {
		logger.Error("lack of user  information for update ")
		RenderJSON(res, req, RetUser{
			RetCode: -21,
			Msg:     "update error with no data"})
		return
	}

	olderuserbytes, err := db.User(email)
	if err != nil {
		logger.Error("Update not exists user")
		RenderJSON(res, req, RetUser{
			RetCode: -1,
			Msg:     "update non-exsit user"})
		return
	}
	ouser := user.User{}
	json.Unmarshal(olderuserbytes, &ouser)

	var updatedUser *user.User
	if len(password) > 0 {
		updatedUser, err = user.New(email, password)
		if err != nil {
			logger.Error("password error")
			RenderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     "update "})
			return
		}
	} else {
		updatedUser = new(user.User)
		logger.Debug("password no to change")
	}
	updatedUser.ID = ouser.ID
	updatedUser.Email = email
	if len(social) > 0 {
		updatedUser.Meta = meta
		updatedUser.Social = social
	}
	err = db.UpdateUser(&ouser, updatedUser)

	if err != nil {
		logger.Error(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	ipSearchHandler := ip.NewClient("", true)

	countryInfor, _ := ipSearchHandler.QueryIPByDB(ipAddr)
	week := time.Now().Add(time.Hour * 24 * 7)
	claims := map[string]interface{}{
		"exp":     week,
		"user":    updatedUser.Email,
		"country": countryInfor,
	}
	token, err := jwt.New(claims)

	if err != nil {
		logger.Debug(err)
		RenderJSON(res, req, RetUser{
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

	currency := getContentList("Currency")
	country := getContentList("Country")
	button := getContentList("PaymentButton")
	RenderJSON(res, req, RetUser{
		RetCode:        0,
		Msg:            "Done",
		Data:           token,
		DefaultCountry: countryInfor,
		Country:        country,
		Currency:       currency,
		SocialType:     updatedUser.Meta,
		SocialLink:     updatedUser.Social,
		Buttons:        button,
	})
	/*
		RenderJSON(res, req, RetUser{
			RetCode: 0,
			Msg:     "Done",
			Data:    updatedUser.ID, // return user id
		}) */
	return
}

// customer login function , check login credential and return data
func Login(res http.ResponseWriter, req *http.Request) {
	ipAddr := GetIP(req)
	logger.Debugf("User login, from:", ipAddr)
	if user.IsValid(req) {
		logger.Debug("is valid")
		//http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		RenderJSON(res, req,
			RetUser{
				RetCode: 2,
				Msg:     "Already Loggin",
				Data:    "",
			})
		return
	}
	requestJson := GetJsonFromBody(req)

	// check email & password

	email := fmt.Sprintf("%s", requestJson["email"])
	if email == "" {
		logger.Error("No login infor!")
		RenderJSON(res, req,
			RetUser{
				RetCode: -2,
				Msg:     "No Login Information",
			})
		return
	}
	password := fmt.Sprintf("%s", requestJson["password"])
	logger.Debug("The Request email is :", email)

	j, err := db.User(strings.ToLower(email))

	if err != nil {
		logger.Error(err)
		RenderJSON(res, req, RetUser{
			RetCode: -1,
			Msg:     err.Error(),
			Data:    "",
		})
		return
	}

	if j == nil {
		logger.Error("no such user")
		RenderJSON(res, req, RetUser{
			RetCode: -1,
			Msg:     "no such user"})
		return
	}

	usr := &user.User{}
	err = json.Unmarshal(j, usr)
	if err != nil {
		logger.Error(err)
		RenderJSON(res, req, RetUser{
			RetCode: -1,
			Msg:     err.Error(),
			Data:    "",
		})
		return
	}

	if !user.IsUser(usr, password) { //check if user password is ok
		logger.Warn("wrong user login attempt")
		RenderJSON(res, req, RetUser{
			RetCode: -1,
			Msg:     "Wrong email or password!",
			Data:    "",
		})
		return
	}
	// create new token
	week := time.Now().Add(time.Hour * 2) // session time is 2 hours

	ipSearchHandler := ip.NewClient("", true)

	countryInfor, _ := ipSearchHandler.QueryIPByDB(ipAddr)

	claims := map[string]interface{}{
		"exp":     week,
		"user":    usr.Email,
		"country": countryInfor,
	}
	token, err := jwt.New(claims)
	//DecodeJwt(token)
	//logger.Debug(jwt.GetClaims(token))

	if err != nil {
		logger.Debug(err)
		RenderJSON(res, req, RetUser{
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
	currency := getContentList("Currency")
	country := getContentList("Country")
	button := getContentList("PaymentButton")
	RenderJSON(res, req, RetUser{
		RetCode:        0,
		Msg:            "Done",
		Data:           token,
		DefaultCountry: countryInfor,
		Country:        country,
		Currency:       currency,
		SocialType:     usr.Meta,
		SocialLink:     usr.Social,
		Buttons:        button,
	})

	return

}

// customer login function , check login credential and return data
func Config(res http.ResponseWriter, req *http.Request) {
	ipAddr := GetIP(req)
	logger.Debugf("User access web site , from:%s", ipAddr)

	ipSearchHandler := ip.NewClient("", true)

	countryInfor, _ := ipSearchHandler.QueryIPByDB(ipAddr)

	currency := getContentList("Currency")
	country := getContentList("Country")
	RenderJSON(res, req, RetUser{
		RetCode:        0,
		Msg:            "Done",
		Data:           "",
		DefaultCountry: countryInfor,
		Country:        country,
		Currency:       currency,
		SocialType:     "",
		SocialLink:     "",
	})
	return
}

func Renew(res http.ResponseWriter, req *http.Request) {
	logger.Debugf("User try to renew token ,From :", GetIP(req))

	if user.IsValid(req) {
		logger.Debug("is valid")
		//http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		//week := time.Now().Add(time.Hour * 2) // session time is 2 hours
		// check if token exists in cookie
		cookie, err := req.Cookie(user.Lqcmstoken)
		if err == nil {
			// validate it and allow or redirect request
			token := cookie.Value
			claim := jwt.GetClaims(token)

			// create new token
			week := time.Now().Add(time.Hour * 2) // session time is 2 hours

			claims := map[string]interface{}{
				"exp":  week,
				"user": claim["user"],
			}
			newtoken, err := jwt.New(claims)
			//DecodeJwt(token)
			//logger.Debug(jwt.GetClaims(token))

			if err != nil {
				logger.Debug(err)
				RenderJSON(res, req, RetUser{
					RetCode: -1,
					Msg:     "Internal Error",
					Data:    "",
				})
				return
			}

			http.SetCookie(res, &http.Cookie{
				Name:    user.Lqcmstoken,
				Value:   newtoken,
				Expires: week,
				Path:    "/",
			})

			RenderJSON(res, req, RetUser{
				RetCode: 0,
				Msg:     "Done",
				Data:    newtoken,
			})
			logger.Debug("User renew !")
		} else {
			logger.Error("no token when renew ")
			RenderJSON(res, req, RetUser{
				RetCode: -10,
				Msg:     err.Error(),
			})
		}
	} else {
		logger.Error("Not valide user try to renew token")
		RenderJSON(res, req, RetUser{
			RetCode: -9,
			Msg:     "not valid user",
		})
	}
	return

}

func Logout(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:    user.Lqcmstoken,
		Expires: time.Unix(0, 0),
		Value:   "",
		Path:    "/",
	})
	RenderJSON(res, req, RetUser{
		RetCode: 0,
		Msg:     "Done",
		Data:    "",
	})
	return
	//	http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/login", http.StatusFound)
}

// allow user to request recover key
func Forgot(res http.ResponseWriter, req *http.Request) {
	logger.Debugf("User try to recover password, from:", GetIP(req))
	reqJSON := GetJsonFromBody(req)

	// check email for user, if no user return Error
	email := strings.ToLower(fmt.Sprintf("%s", reqJSON["email"]))
	if email == "" {
		res.WriteHeader(http.StatusBadRequest)
		logger.Error("Failed account recovery. No email address submitted.")
		return
	}

	_, err = db.User(email)
	if err == db.ErrNoUserExists {
		res.WriteHeader(http.StatusBadRequest)
		logger.Error("No user exists.", err)
		return
	}

	if err != db.ErrNoUserExists && err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		logger.Error("Error:", err)
		return
	}

	// create temporary key to verify user
	key, err := db.SetRecoveryKey(email)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		logger.Error("Failed to set account recovery key.", err)
		return
	}

	domain, err := db.Config("domain")
	/* emailhost, err := db.Config("email_host")
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		logger.Error("Failed to get domain from configuration.", err)
		return
	}
	emailsecret, err := db.Config("email_password") */
	/* adminemail, err := db.Config("admin_email")
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		logger.Debugf("Please set admin email box to send recover letter.", err)
		return
	} */
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

	/* msg := emailer.Message{
		To:      email,
		From:    fmt.Sprintf("admin@%s", domain),
		Subject: fmt.Sprintf("Account Recovery [%s]", domain),
		Body:    body,
	} */

	go func() {
		//err = msg.Send()
		tomail := []string{email}

		fmt.Printf("Try to send  notification email to %v\n", tomail)
		smtpemail := smtp2go.Email{
			//From: admin.MailUser,
			To:       tomail,
			Subject:  fmt.Sprintf("Account Recovery [%s]", "EGPal"),
			TextBody: body,
			HtmlBody: body,
		}
		res, err := smtp2go.Send(&smtpemail)
		if err != nil {
			fmt.Printf("Send Alert email with n Error Occurred: %s\n", err)
		}
		if res.Data.Succeeded == 1 {
			fmt.Printf("Email allter sent Successfully: %v\n", res)
		} else {
			fmt.Printf("Email allter Sent with error: %v\n", res)
		}
		/* err = SendEmail(string(emailhost[:]), string(adminemail[:]), email, string(emailsecret[:]), fmt.Sprintf("Account Recovery [%s]", "恩卓信息"), body)

		if err != nil {
			logger.Debugf("Failed to send message to:", email, "Error:", err)
		} else {
			logger.Debug("Recover email sent out without error  to ", email)
		} */
	}()

	// redirect to /admin/recover/key and send email with key and URL
	//http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/recover/key", http.StatusFound)
	RenderJSON(res, req, RetUser{RetCode: 0, Msg: "Recovery Email sent, Please check!"})

}

func Recovery(res http.ResponseWriter, req *http.Request) {
	logger.Debugf("User try to recover password form :%s", GetIP(req))
	reqJSON := GetJsonFromBody(req)

	// check for email & key match
	email := strings.ToLower(fmt.Sprintf("%s", reqJSON["email"]))
	key := fmt.Sprintf("%s", reqJSON["key"])

	if email == "" || key == "" {
		res.WriteHeader(http.StatusBadRequest)
		logger.Debug("Failed account recovery. No email address submitted.")
		return
	}
	var actual string
	if actual, err = db.RecoveryKey(email); err != nil || actual == "" {
		logger.Debug("Error getting recovery key from database:", err)
		RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again."})
		return
	}

	if key != actual {
		logger.Debug("Bad recovery key submitted:", key)

		RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
		return
	}

	// set user with new password
	password := fmt.Sprintf("%s", reqJSON["password"])
	usr := &user.User{}
	u, err := db.User(email)
	if err != nil {
		logger.Error("Error finding user by email:", email, err)

		RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
		return
	}

	if u == nil {
		logger.Error("No user found with email:", email)

		RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error,  please go back and try again.", Data: ""})
		return
	}

	err = json.Unmarshal(u, usr)
	if err != nil {
		logger.Error("Error decoding user from database:", err)

		RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
		return
	}

	update, err := user.NewCustomer(email, password)
	if err != nil {
		logger.Error(err)

		RenderJSON(res, req, RetUser{RetCode: -1, Msg: err.Error(), Data: ""})
		return
	}

	update.ID = usr.ID

	err = db.UpdateUser(usr, update)
	if err != nil {
		logger.Error("Error updating user:", err)
		RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
		return
	}

	RenderJSON(res, req, RetUser{RetCode: 1, Msg: "Done,Pleaes relogin", Data: usr.ID})
	logger.Debugf("User %s recover password, Done", usr.Email)
}

func LogoutHandler(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:    user.Lqcmstoken,
		Expires: time.Unix(0, 0),
		Value:   "",
		Path:    "/",
	})
	RenderJSON(res, req, RetUser{
		RetCode: 0,
		Msg:     "Done",
		Data:    "",
	})
	return
	//	http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/login", http.StatusFound)
}

func LoginHandler(res http.ResponseWriter, req *http.Request) {
	logger.Debugf("User login, from:", GetIP(req))
	switch req.Method {

	case http.MethodPost:
		if user.IsValid(req) {
			logger.Debug("is valid")
			//http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			RenderJSON(res, req,
				RetUser{
					RetCode: 2,
					Msg:     "Already Loggin",
					Data:    "",
				})
			return
		}
		requestJson := GetJsonFromBody(req)
		//fmt.Println(requestJson)
		/* err := req.ParseForm()

		if err != nil {
			log.Println(err)
			RenderJSON(res, req, RetUser{
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
			RenderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     err.Error(),
				Data:    "",
			})
			return
		}

		if j == nil {
			RenderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     "no such user"})
			return
		}

		usr := &user.User{}
		err = json.Unmarshal(j, usr)
		if err != nil {
			log.Println(err)
			RenderJSON(res, req, RetUser{
				RetCode: -1,
				Msg:     err.Error(),
				Data:    "",
			})
			return
		}

		if !user.IsUser(usr, password) {
			RenderJSON(res, req, RetUser{
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
		//logger.Debug(jwt.GetClaims(token))

		if err != nil {
			logger.Error(err)
			RenderJSON(res, req, RetUser{
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
		RenderJSON(res, req, RetUser{
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
				RenderJSON(res, req, RetUser{
					RetCode: 0,
					Msg:     "Done",
					Data:    "",
				})
			} else {
				RenderJSON(res, req, RetUser{
					RetCode: -10,
					Msg:     err.Error(),
				})
			}
		} else {
			RenderJSON(res, req, RetUser{
				RetCode: -9,
				Msg:     "not valid user",
			})
		}
		return
	default:
		//		logger.Debugf("User %s logged in !", usr)
		RenderJSON(res, req, RetUser{
			RetCode: -10,
			Msg:     "failed",
			Data:    "",
		})

		return
	}
}

func ForgotPasswordHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			RenderJSON(res, req, RetUser{
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
		/*
			msg := emailer.Message{
				To:      email,
				From:    fmt.Sprintf("admin@%s", domain),
				Subject: fmt.Sprintf("Account Recovery [%s]", domain),
				Body:    body,
			} */

		go func() {
			tomail := []string{email}

			fmt.Printf("Try to send admin notification email to %v\n", tomail)

			emailTarget := smtp2go.Email{
				//From: admin.MailUser,
				To:       tomail,
				Subject:  fmt.Sprintf("Account Recovery [%s]", "EGPal"),
				TextBody: body,
				HtmlBody: body,
			}
			res, err := smtp2go.Send(&emailTarget)
			if err != nil {
				fmt.Printf("Send Alert email with n Error Occurred: %s\n", err)
			}
			if res.Data.Succeeded == 1 {
				fmt.Printf("Email allter sent Successfully: %v\n", res)
			} else {
				fmt.Printf("Email allter Sent with error: %v\n", res)
			}
			/*
				err = msg.Send()
				if err != nil {
					log.Println("Failed to send message to:", msg.To, "about", msg.Subject, "Error:", err)
				} */
		}()

		// redirect to /admin/recover/key and send email with key and URL
		//http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/recover/key", http.StatusFound)
		RenderJSON(res, req, RetUser{RetCode: 0, Msg: "Done"})
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

			RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again."})
			return
		}

		// check for email & key match
		email := strings.ToLower(req.FormValue("email"))
		key := req.FormValue("key")

		var actual string
		if actual, err = db.RecoveryKey(email); err != nil || actual == "" {
			log.Println("Error getting recovery key from database:", err)
			RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again."})
			return
		}

		if key != actual {
			log.Println("Bad recovery key submitted:", key)

			RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
			return
		}

		// set user with new password
		password := req.FormValue("password")
		usr := &user.User{}
		u, err := db.User(email)
		if err != nil {
			log.Println("Error finding user by email:", email, err)

			RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
			return
		}

		if u == nil {
			log.Println("No user found with email:", email)

			RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error,  please go back and try again.", Data: ""})
			return
		}

		err = json.Unmarshal(u, usr)
		if err != nil {
			log.Println("Error decoding user from database:", err)

			RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
			return
		}

		update, err := user.NewCustomer(email, password)
		if err != nil {
			log.Println(err)

			RenderJSON(res, req, RetUser{RetCode: -1, Msg: err.Error(), Data: ""})
			return
		}

		update.ID = usr.ID

		err = db.UpdateUser(usr, update)
		if err != nil {
			log.Println("Error updating user:", err)
			RenderJSON(res, req, RetUser{RetCode: -1, Msg: "Error, please go back and try again.", Data: ""})
			return
		}

		RenderJSON(res, req, RetUser{RetCode: 1, Msg: "Done,Pleaes relogin", Data: ""})

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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
