package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/agreyfox/eshop/management/editor"
	"github.com/agreyfox/eshop/management/format"
	"github.com/agreyfox/eshop/management/manager"
	"github.com/agreyfox/eshop/system/addon"
	"github.com/agreyfox/eshop/system/admin/config"
	"github.com/agreyfox/eshop/system/admin/upload"
	"github.com/agreyfox/eshop/system/admin/user"
	"github.com/agreyfox/eshop/system/api"
	"github.com/agreyfox/eshop/system/api/analytics"
	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/item"
	"github.com/agreyfox/eshop/system/search"
	"github.com/nfnt/resize"

	"github.com/gorilla/schema"
	emailer "github.com/nilslice/email"
	"github.com/nilslice/jwt"
)

func login(res http.ResponseWriter, req *http.Request) {
	logger.Debug("login in request from:", GetIP(req))
	if user.IsValid(req) {
		renderJSON(res, req, ReturnData{
			RetCode: 2,
			Msg:     "Already Login",
		})
		return
	}
	requestJson := getJsonFromBody(req)
	if requestJson == nil {
		renderJSON(res, req, ReturnData{
			RetCode: -1,
			Msg:     "No User Input",
		})
		return
	}
	email := requestJson["email"].(string)
	password := requestJson["password"].(string)

	// check email & password
	j, err := db.User(strings.ToLower(email))
	if err != nil || j == nil {
		logger.Error(err)
		renderJSON(res, req, ReturnData{
			RetCode: -1,
			Msg:     "No User information",
		})
		return
	}

	usr := &user.User{}
	err = json.Unmarshal(j, usr)
	if err != nil {
		logger.Error(err)
		renderJSON(res, req, ReturnData{
			RetCode: -1,
			Msg:     err.Error(),
		})
		return
	}

	if !user.IsUser(usr, password) {
		renderJSON(res, req, ReturnData{
			RetCode: -99,
			Msg:     "user name or password incorrect",
		})
		return
	}
	if !usr.Perm.Admin {
		logger.Warnf("Normal user try to access admin panel")
		renderJSON(res, req, ReturnData{
			RetCode: -99,
			Msg:     "Permission Denied",
		})
		return
	}
	// create new token
	week := time.Now().Add(time.Hour * 24 * 7)
	claims := map[string]interface{}{
		"exp":  week,
		"user": usr.Email,
	}
	token, err := jwt.New(claims)
	if err != nil {
		logger.Error(err)
		renderJSON(res, req, ReturnData{
			RetCode: -6,
			Msg:     "Internal Error",
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
	contentStructData := getContentsStruct()
	logger.Debugf("Admin User %s logged in !", usr.Email)
	retdata := map[string]interface{}{
		"retCode":  0,
		"msg":      "Done",
		"data":     token,
		"contents": string(contentStructData[:]),
	}

	renderJSON(res, req, retdata)
	return
}

func logout(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:    user.Lqcmstoken,
		Expires: time.Unix(0, 0),
		Value:   "",
		Path:    "/",
	})
	renderJSON(res, req, ReturnData{
		RetCode: 0,
		Msg:     "Done",
	})
	http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/", http.StatusFound)
}

func recoverRequest(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("User try to recover password form :%s", GetIP(r))

	reqJSON := getJsonFromBody(r)
	// check email for user, if no user return Error
	email := strings.ToLower(fmt.Sprintf("%s", reqJSON["email"]))
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger.Debug("Failed account recovery. No email address submitted.")
		return
	}

	_, err = db.User(email)
	if err == db.ErrNoUserExists {
		w.WriteHeader(http.StatusBadRequest)
		logger.Warn("No user exists.", err)
		return
	}

	if err != db.ErrNoUserExists && err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Debug("Error:", err)
		return
	}

	// create temporary key to verify user
	key, err := db.SetRecoveryKey(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Debugf("Failed to set account recovery key.", err)
		return
	}

	domain, err := db.Config("domain")
	emailhost, err := db.Config("email_host")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Debugf("Failed to get domain from configuration.", err)
		return
	}
	emailsecret, err := db.Config("email_password")
	adminemail, err := db.Config("admin_email")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Debugf("Please set admin email box to send recover letter.", err)
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
at %s

`, email, domain, key, domain)
	/*
		msg := emailer.Message{
			To:      email,
			From:    fmt.Sprintf("e_raeb@%s", domain),
			Subject: fmt.Sprintf("Account Recovery [%s]", domain),
			Body:    body,
		} */

	go func() {
		//err = msg.Send()
		err = sendEmail(string(emailhost[:]), string(adminemail[:]), email, string(emailsecret[:]), fmt.Sprintf("Account Recovery [%s]", "恩卓信息"), body)
		if err != nil {
			logger.Debugf("Failed to send message to:", email, "Error:", err)
		} else {
			logger.Debug("Recover email sent out without error  to ", email)
		}
	}()

	renderJSON(w, r, map[string]interface{}{
		"retCode": 0,
		"msg":     "Recovery Email sent, Please check ",
	})

}

func recoverPassword(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("User try to recover password form :%s", GetIP(r))
	reqJSON := getJsonFromBody(r)
	// check email for user, if no user return Error
	email := strings.ToLower(fmt.Sprintf("%s", reqJSON["email"]))
	key := strings.ToLower(fmt.Sprintf("%s", reqJSON["key"].(string)))
	if email == "" || key == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger.Debug("Failed account recovery. No email address submitted.")
		return
	}

	var actual string
	if actual, err = db.RecoveryKey(email); err != nil || actual == "" {
		logger.Error("Error getting recovery key from database:", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error, please go back and try again."))
		return
	}

	if key != actual {
		logger.Debug("Bad recovery key submitted:", key)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error, please go back and try again."))
		return
	}

	// set user with new password
	password := fmt.Sprintf("%s", reqJSON["password"])
	usr := &user.User{}
	u, err := db.User(email)
	if err != nil {
		logger.Debug("Error finding user by email:", email, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error, please go back and try again."))
		return
	}

	if u == nil {
		logger.Debug("No user found with email:", email)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error, please go back and try again."))
		return
	}

	err = json.Unmarshal(u, usr)
	if err != nil {
		logger.Debugf("Error decoding user from database:", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error, please go back and try again."))
		return
	}

	update, err := user.New(email, password)
	if err != nil {
		logger.Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error, please go back and try again."))
		return
	}

	update.ID = usr.ID

	err = db.UpdateUser(usr, update)
	if err != nil {
		logger.Debug("Error updating user:", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error, please go back and try again."))
		return
	}
	renderJSON(w, r, map[string]interface{}{
		"retCode": 0,
		"msg":     "done,Please login with new password",
	})
	logger.Debugf("User %s recover password, Done", usr.Email)
}

//get all changeable config
func getConfig(res http.ResponseWriter, req *http.Request) {
	data, err := db.ConfigAll()
	if err != nil {
		logger.Error(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := &config.Config{}

	err = json.Unmarshal(data, c)

	if err != nil {
		logger.Error(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	ret := map[string]interface{}{}
	ret["name"] = c.Name
	ret["bind_addr"] = c.BindAddress
	ret["http_port"] = c.HTTPPort
	ret["https_port"] = c.HTTPSPort

	ret["admin_email"] = c.AdminEmail
	ret["cors_disabled"] = c.DisableCORS
	ret["cache_disabled"] = c.DisableHTTPCache
	ret["zip_disabled"] = c.DisableGZIP
	ret["log_level"] = c.LogLevel
	ret["log_file"] = c.LogFile
	ret["domain"] = c.Domain
	ret["email_password"] = c.EmailSecret
	ret["email_host"] = c.EmailHost
	//output all the config to req
	returnStructData(res, req, []map[string]interface{}{{"data": ret}}, MetaData{})

}

//save the config to system, could be key-value and multiple is support
func saveConfig(res http.ResponseWriter, req *http.Request) {
	logger.Debugf("Admin try to save the system configuration,from", GetIP(req))
	if user.IsValidAdmin(req) {
		//fmt.Println("is admin")
		ret := getJsonFromBody(req)
		if ret == nil {
			renderJSON(res, req, ReturnData{
				RetCode: -1,
				Msg:     "No Input Data",
			})
			return
		}
		//PrettyPrint(ret)
		var err error
		var someerror bool
		for k, v := range ret {
			err = db.PutConfig(k, v)
			if err != nil {
				logger.Errorf("Save key %s error", k)
				someerror = true
			}
		}

		/* 		{
		   			"name": ret["name"].(string),

		   		"bind_addr" : ret["bind_addr"].(string),
		   		"http_port": ret["http_port"].(string)
		   		c.HTTPSPort : ret["https_port"].(string)

		   		c.AdminEmail : ret["admin_email"].(string)
		   		c.DisableCORS : ret["cors_disabled"].(bool)
		   		c.DisableHTTPCache : ret["cache_disabled"].(bool)
		   		c.DisableGZIP : ret["zip_disabled"].(bool)
		   		c.LogLevel : ret["log_level"].(string)
		   		c.LogFile : ret["log_file"].(string)

		   		} */

		if someerror {
			renderJSON(res, req, ReturnData{
				RetCode: 0,
				Msg:     "Some config save error ",
			})
			return
		}

		renderJSON(res, req, ReturnData{
			RetCode: 0,
			Msg:     "Done",
		})
	} else {
		renderJSON(res, req, ReturnData{
			RetCode: -99,
			Msg:     "Permission Denied",
		})
	}
}

func configRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {

	// /admin/
	case http.MethodGet:
		data, err := db.ConfigAll()
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		c := &config.Config{}

		err = json.Unmarshal(data, c)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		//output all the config to req
		renderJSON(res, req, data)

	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = db.SetConfig(req.Form)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(res, req, req.URL.String(), http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func backup(res http.ResponseWriter, req *http.Request) {
	logger.Debug("Admin try to backup system ,from :", GetIP(req))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch req.URL.Query().Get("source") {
	case "system":
		err := db.Backup(ctx, res)
		if err != nil {
			logger.Error("Failed to run backup on system:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "analytics":
		err := analytics.Backup(ctx, res)
		if err != nil {
			logger.Error("Failed to run backup on analytics:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "uploads":
		err := upload.Backup(ctx, res)
		if err != nil {
			logger.Error("Failed to run backup on uploads:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "search":
		err := search.Backup(ctx, res)
		if err != nil {
			logger.Error("Failed to run backup on search:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		res.WriteHeader(http.StatusBadRequest)
	}
}

func backupRestHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch req.URL.Query().Get("source") {
	case "system":
		err := db.Backup(ctx, res)
		if err != nil {
			log.Println("Failed to run backup on system:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "analytics":
		err := analytics.Backup(ctx, res)
		if err != nil {
			log.Println("Failed to run backup on analytics:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "uploads":
		err := upload.Backup(ctx, res)
		if err != nil {
			log.Println("Failed to run backup on uploads:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "search":
		err := search.Backup(ctx, res)
		if err != nil {
			log.Println("Failed to run backup on search:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		res.WriteHeader(http.StatusBadRequest)
	}
}

func configUsersRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		view, err := UsersList(req)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		res.Write(view)

	case http.MethodPost:
		// create new user
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		email := strings.ToLower(req.FormValue("email"))
		password := req.PostFormValue("password")

		if email == "" || password == "" {
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		usr, err := user.New(email, password)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = db.SetUser(usr)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		http.Redirect(res, req, req.URL.String(), http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func configUsersEditRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// check if user to be edited is current user
		j, err := db.CurrentUser(req)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		usr := &user.User{}
		err = json.Unmarshal(j, usr)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// check if password matches
		password := req.PostFormValue("password")

		if !user.IsUser(usr, password) {
			log.Println("Unexpected user/password combination for", usr.Email)
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error405()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		email := strings.ToLower(req.PostFormValue("email"))
		newPassword := req.PostFormValue("new_password")
		var updatedUser *user.User
		if newPassword != "" {
			updatedUser, err = user.New(email, newPassword)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			updatedUser, err = user.New(email, password)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// set the ID to the same ID as current user
		updatedUser.ID = usr.ID

		// set user in db
		err = db.UpdateUser(usr, updatedUser)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// create new token
		week := time.Now().Add(time.Hour * 24 * 7)
		claims := map[string]interface{}{
			"exp":  week,
			"user": updatedUser.Email,
		}
		token, err := jwt.New(claims)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// add token to cookie +1 week expiration
		cookie := &http.Cookie{
			Name:    user.Lqcmstoken,
			Value:   token,
			Expires: week,
			Path:    "/",
		}
		http.SetCookie(res, cookie)

		// add new token cookie to the request
		req.AddCookie(cookie)

		http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/edit"), http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func configUsersDeleteRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// do not allow current user to delete themselves
		j, err := db.CurrentUser(req)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		usr := &user.User{}
		err = json.Unmarshal(j, &usr)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		email := strings.ToLower(req.PostFormValue("email"))

		if usr.Email == email {
			log.Println(err)
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error405()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// delete existing user
		err = db.DeleteUser(email)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/delete"), http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func loginRestHandler(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:
		if user.IsValid(req) {
			http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			return
		}

		view, err := Login()
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(view)

	case http.MethodPost:
		logger.Debug("login in request from:", GetIP(req))
		if user.IsValid(req) {
			http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			return
		}

		err := req.ParseForm()
		//logger.Debugf("%v", req)
		if err != nil {
			log.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		// check email & password
		j, err := db.User(strings.ToLower(req.FormValue("email")))
		if err != nil {
			log.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		if j == nil {
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		usr := &user.User{}
		err = json.Unmarshal(j, usr)
		if err != nil {
			log.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		if !user.IsUser(usr, req.FormValue("password")) {
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}
		// create new token
		week := time.Now().Add(time.Hour * 24 * 7)
		claims := map[string]interface{}{
			"exp":  week,
			"user": usr.Email,
		}
		token, err := jwt.New(claims)
		if err != nil {
			log.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		// add it to cookie +1 week expiration
		http.SetCookie(res, &http.Cookie{
			Name:    user.Lqcmstoken,
			Value:   token,
			Expires: week,
			Path:    "/",
		})

		http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/login"), http.StatusFound)
	}
}

func logoutRestHandler(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:    user.Lqcmstoken,
		Expires: time.Unix(0, 0),
		Value:   "",
		Path:    "/",
	})

	http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/login", http.StatusFound)
}

func forgotPasswordRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		view, err := ForgotPassword()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		res.Write(view)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
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
Dms at %s

`, email, domain, key, domain)

		msg := emailer.Message{
			To:      email,
			From:    fmt.Sprintf("dms@%s", domain),
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
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/recover/key", http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		errView, err := Error405()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}
}

func recoveryKeyRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		view, err := RecoveryKey()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Write(view)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println("Error parsing recovery key form:", err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		// check for email & key match
		email := strings.ToLower(req.FormValue("email"))
		key := req.FormValue("key")

		var actual string
		if actual, err = db.RecoveryKey(email); err != nil || actual == "" {
			log.Println("Error getting recovery key from database:", err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		if key != actual {
			log.Println("Bad recovery key submitted:", key)

			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		// set user with new password
		password := req.FormValue("password")
		usr := &user.User{}
		u, err := db.User(email)
		if err != nil {
			log.Println("Error finding user by email:", email, err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		if u == nil {
			log.Println("No user found with email:", email)

			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		err = json.Unmarshal(u, usr)
		if err != nil {
			log.Println("Error decoding user from database:", err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		update, err := user.New(email, password)
		if err != nil {
			log.Println(err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		update.ID = usr.ID

		err = db.UpdateUser(usr, update)
		if err != nil {
			log.Println("Error updating user:", err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		// redirect to /admin/login
		redir := req.URL.Scheme + req.URL.Host + "/admin/login"
		http.Redirect(res, req, redir, http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func uploadContentsRestHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()

	order := strings.ToLower(q.Get("order"))
	if order != "asc" {
		order = "desc"
	}

	pt := interface{}(&item.FileUpload{})

	p, ok := pt.(editor.Editable)
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if q.Get("count") == "" {
			count = 10
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}
	}

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}
	}

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}

	b := &bytes.Buffer{}
	var total int
	var posts [][]byte

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="col s8">
						<div class="row">
							<div class="card-title col s7">Uploaded Items</div>
							<div class="col s5 input-field inline">
								<select class="browser-default __dms sort-order">
									<option value="DESC">New to Old</option>
									<option value="ASC">Old to New</option>
								</select>
								<label class="active">Sort:</label>
							</div>	
							<script>
								$(function() {
									var sort = $('select.__dms.sort-order');

									sort.on('change', function() {
										var path = window.location.pathname;
										var s = sort.val();

										window.location.replace(path + '?order=' + s);
									});

									var order = getParam('order');
									if (order !== '') {
										sort.val(order);
									}
									
								});
							</script>
						</div>
					</div>
					<form class="col s4" action="/admin/uploads/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all Upload fields" class="search"/>
							<input type="hidden" name="type" value="DB__uploads" />
						</div>
                    </form>	
					</div>`

	t := db.DB__uploads // upload db
	status := ""
	total, posts = db.Query(t, opts)

	for i := range posts {
		err := json.Unmarshal(posts[i], &p)
		if err != nil {
			log.Println("Error unmarshal json into", t, err, string(posts[i]))

			post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
			_, err := b.Write([]byte(post))
			if err != nil {
				log.Println(err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					log.Println(err)
				}

				res.Write(errView)
				return
			}
			continue
		}

		post := adminPostListItem(p, t, status)
		_, err = b.Write(post)
		if err != nil {
			log.Println(err)

			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				log.Println(err)
			}

			res.Write(errView)
			return
		}
	}

	html += `<ul class="posts row">`

	_, err = b.Write([]byte(`</ul>`))
	if err != nil {
		log.Println(err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			log.Println(err)
		}

		res.Write(errView)
		return
	}

	statusDisabled := "disabled"
	prevStatus := ""
	nextStatus := ""
	// total may be less than 10 (default count), so reset count to match total
	if total < count {
		count = total
	}
	// nothing previous to current list
	if offset == 0 {
		prevStatus = statusDisabled
	}
	// nothing after current list
	if (offset+1)*count >= total {
		nextStatus = statusDisabled
	}

	// set up pagination values
	urlFmt := req.URL.Path + "?count=%d&offset=%d&&order=%s"
	prevURL := fmt.Sprintf(urlFmt, count, offset-1, order)
	nextURL := fmt.Sprintf(urlFmt, count, offset+1, order)
	start := 1 + count*offset
	end := start + count - 1

	if total < end {
		end = total
	}

	pagination := fmt.Sprintf(`
	<ul class="pagination row">
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_left</i></a></li>
		<li class="col s8">%d to %d of %d</li>
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_right</i></a></li>
	</ul>
	`, prevStatus, prevURL, start, end, total, nextStatus, nextURL)

	// show indicator that a collection of items will be listed implicitly, but
	// that none are created yet
	if total < 1 {
		pagination = `
		<ul class="pagination row">
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_left</i></a></li>
			<li class="col s8">0 to 0 of 0</li>
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_right</i></a></li>
		</ul>
		`
	}

	_, err = b.Write([]byte(pagination + `</div></div>`))
	if err != nil {
		log.Println(err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			log.Println(err)
		}

		res.Write(errView)
		return
	}

	script := `
	<script>
		$(function() {
			var del = $('.quick-delete-post.__dms span');
			del.on('click', function(e) {
				if (confirm("[Dms] Please confirm:\n\nAre you sure you want to delete this post?\nThis cannot be undone.")) {
					$(e.target).parent().submit();
				}
			});
		});

		// disable link from being clicked if parent is 'disabled'
		$(function() {
			$('ul.pagination li.disabled a').on('click', function(e) {
				e.preventDefault();
			});
		});
	</script>
	`

	btn := `<div class="col s3"><a href="/admin/edit/upload" class="btn new-post waves-effect waves-light">New Upload</a></div></div>`
	html = html + b.String() + script + btn

	adminView, err := Admin([]byte(html))
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}

//get content attachement
func getMediaContents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	order := strings.ToLower(q.Get("order"))
	if order != "asc" {
		order = "desc"
	}

	//	pt := interface{}(&item.FileUpload{})

	/* 	p, ok := pt.(editor.Editable)
	   	if !ok {
	   		w.WriteHeader(http.StatusInternalServerError)

	   		return
	   	}
	*/
	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if q.Get("count") == "" {
			count = 10
		} else {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	}

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
		} else {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	}

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}

	//b := &bytes.Buffer{}
	var total int
	var posts [][]byte

	//html := `uploads`

	t := db.DB__uploads // upload db
	//status := ""
	total, posts = db.Query(t, opts)
	retData := make([]map[string]interface{}, 0)

	for i := range posts {
		item := make(map[string]interface{})
		err := json.Unmarshal(posts[i], &item)
		if err != nil {
			logger.Error("Error unmarshal json into", t, err, string(posts[i]))

			continue
		}
		retData = append(retData, item)

	}

	paa := int(0)

	if count > 0 {
		paa = total / count
		if paa == 0 {
			paa = 1
		}
	} else {
		paa = 1
	}

	meta := MetaData{
		Total:     uint(total),
		PageCount: paa,
		Page:      offset,
		Order:     order,
		PageSize:  count, //-1 means all
	}

	returnStructData(w, r, retData, meta)

	logger.Debugf("get all media library list ,total %d record", total)

}

// to search media content
func searchMediaContent(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Search Media library from :", GetIP(r))
	q := r.URL.Query()
	t := db.DB__uploads
	search := q.Get("q")
	//status := q.Get("status")

	if t == "" || search == "" {
		logger.Error("Search parameter is not proper")
		w.WriteHeader(http.StatusBadRequest)
		//	http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	posts := db.UploadAll()
	total := len(posts)
	retData := make([]map[string]interface{}, 0)

	for i := range posts {
		// skip posts that don't have any matching search criteria
		match := strings.ToLower(search)
		all := strings.ToLower(string(posts[i]))
		if !strings.Contains(all, match) {
			continue
		}
		item := make(map[string]interface{})
		err := json.Unmarshal(posts[i], &item)
		if err != nil {
			logger.Debug("Error unmarshal search result json into", t, err, posts[i])
			continue
		}
		retData = append(retData, item)
	}

	meta := MetaData{
		Total:     uint(total),
		PageCount: 1,
		Page:      0,
		Order:     "",
		PageSize:  len(retData), //-1 means all
	}
	//fmt.Println(meta)
	returnStructData(w, r, retData, meta)

}

//get media content to show
//？id=xxxxx
func getMedia(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	id := q.Get("id") // int: multiplier of count for pagination (0 default)
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	t := db.DB__uploads // upload db
	//status := ""
	contentbyte, err := db.Upload(t + ":" + id)
	item := make(map[string]interface{})
	//fmt.Println(contentbyte)
	err = json.Unmarshal(contentbyte, &item)
	if err != nil {
		logger.Error("Error unmarshal json into", t, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	fff := item["path"].(string)
	ctype := item["content_type"].(string)
	pwd, err := os.Getwd()
	if err != nil {
		logger.Error("Couldn't find current directory for file server.")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	mediaFilename := filepath.Join(pwd, "uploads", strings.TrimPrefix(fff, "/api/uploads"))
	logger.Debugf("The file  %s being read \n", mediaFilename)
	logger.Debugf(strings.TrimPrefix(fff, "/api/uploads"))
	dat, err := ioutil.ReadFile(mediaFilename)
	if err != nil {
		logger.Error("Couldn't read file content .")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//content_type := mime.TypeByExtension(file_ext)

	logger.Info("content type is %s", ctype)

	if len(ctype) > 0 {
		r.Header.Set("Content-Type", ctype)
	} else {
		r.Header.Set("Content-Type", "image/*")
	}

	width_str := q.Get("w")

	var (
		width        uint64
		is_width_set = false
	)

	if len(width_str) > 0 {

		if width, err = strconv.ParseUint(width_str, 10, 32); nil != err {
			logger.Error("input parameter w is error .", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		is_width_set = true
	}

	height_str := q.Get("h")

	var (
		height        uint64
		is_height_set = false
	)
	if len(height_str) > 0 {

		if height, err = strconv.ParseUint(height_str, 10, 32); nil != err {
			logger.Error("input parameter h is error .", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		is_height_set = true
	}

	logger.Debugf("width and height is [%d, %d], status [%t, %t]", width, height, is_width_set, is_height_set)

	if is_width_set || is_height_set {
		var (
			original_image image.Image
			new_image      image.Image
		)
		if original_image, _, err = image.Decode(bytes.NewReader(dat)); nil != err {
			logger.Error("image decode error! %v", err)
			goto LABEL_IMAGE_HANDLE_FINISHED
		}

		new_image = resize.Resize(uint(width), uint(height), original_image, resize.Lanczos3)
		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, new_image, nil); nil != err {
			logger.Error("image encode error! %v", err)
			goto LABEL_IMAGE_HANDLE_FINISHED
		}
		dat = buf.Bytes()

		r.Header.Set("Content-Type", "image/jpeg")
	}

LABEL_IMAGE_HANDLE_FINISHED:

	w.Write(dat)
	logger.Debugf("send media with id %s", id)

	//http.Redirect(w, r, "/api/uploads"+fmt.Sprint(item["path"]), http.StatusFound)
}

// get conetnt list
func getContents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	t := q.Get("type") //内容 类型
	if t == "" {
		w.WriteHeader(http.StatusBadRequest) // 返回bad request
		logger.Error("Parameter type error")
		return
	}

	logger.Debugf("get Content type %s from :%s", t, GetIP(r))

	order := strings.ToLower(q.Get("order")) //排序
	if order != "asc" {
		order = "desc"
	}

	status := q.Get("status") //状态

	if _, ok := item.Types[t]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("Parameter Status error")
		return
	}

	pt := item.Types[t]()

	/* 	p, ok := pt.(editor.Editable)
	   	if !ok {
	   		w.WriteHeader(http.StatusInternalServerError)
	   		logger.Error("Parameter PT error")
	   		return
	   	}
	*/
	var ok bool
	var hasExt bool
	_, ok = pt.(api.Createable) //创建前用户函数
	if ok {
		hasExt = true
	}
	pageSize := 10
	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err == nil {
		pageSize = count
	}
	//default

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("Parameter Offset error")
			return
		}
	}
	if offset > 0 {
		offset = offset - 1 //match human 1 for page 1,
	}
	opts := db.QueryOptions{ //查询条件个数，页号，排序
		Count:  pageSize,
		Offset: offset,
		Order:  order,
	}

	var specifier string //获取是否已发布，或者等待发布__pending
	if status == "public" || status == "" {
		specifier = "__sorted"
	} else if status == "pending" {
		specifier = "__pending"
	}

	//b := &bytes.Buffer{}
	var total int
	var posts [][]byte
	total = 0
	//html := `html`
	retData := make([]map[string]interface{}, 0)

	if hasExt {
		logger.Debugf("Has ext and ready to run")
		if status == "" {
			q.Set("status", "public") //缺省为public
		}
		switch status {
		case "public", "":
			// get __sorted posts of type t from the db
			total, posts = db.Query(t+specifier, opts) //Look up the db
			//logger.Debugf("%+v", posts)

			for i := range posts {
				item := make(map[string]interface{})

				err := json.Unmarshal(posts[i], &item)

				if err != nil {
					logger.Debugf("Get Data error", err.Error(), i)
					continue
				}
				//suppose to output the list,
				retData = append(retData, item)

			}

			//returnStructData(w, r, retData)

		case "pending":
			// get __pending posts of type t from the db
			logger.Debugf("Querying pending contents ")
			total, posts = db.Query(t+"__pending", opts)
			//logger.Debugf("%+v", posts)

			for i := len(posts) - 1; i >= 0; i-- {
				item := make(map[string]interface{})

				err := json.Unmarshal(posts[i], &item)

				if err != nil {
					logger.Debugf("Get Data error", err.Error(), i)
					continue
				}
				retData = append(retData, item)
			}

		}
	} else {
		logger.Debugf("no creatable item and query public data direclty", opts)
		total, posts = db.Query(t+specifier, opts)

		for i := range posts {
			item := make(map[string]interface{})

			err := json.Unmarshal(posts[i], &item)

			if err != nil {
				logger.Debugf("Get Data error", err.Error(), i)
				continue
			}
			retData = append(retData, item)
			//logger.Debug(retData)
		}

	}
	hook, ok := pt.(item.Hookable)
	if ok {
		// hook before response
		fields, hasSubContent := hook.EnableSubContent()
		if hasSubContent {

			logger.Debug("Now process sub-contents")
			for kk := range retData {
				for index := range fields {
					fieldname := fields[index]
					data, err := db.GetSubContent(t+specifier+":"+fmt.Sprint(retData[kk]["id"]), fieldname)
					//fmt.Println(t + specifier + ":" + fmt.Sprint(retData[kk]["id"]))
					if err == nil {
						outdata := []map[string]interface{}{}
						err := json.Unmarshal(data, &outdata)
						if err == nil {
							retData[kk][fieldname] = outdata
						}
					}

				}
			}

		}

	}

	p := 0

	if pageSize > 0 {
		p = total / pageSize
	} else {
		p = 1
	}

	meta := MetaData{
		Total:     uint(total),
		PageCount: p,
		Page:      offset,
		Order:     order,
		PageSize:  pageSize, //-1 means all
	}
	//fmt.Println(meta)
	returnStructData(w, r, retData, meta)

	logger.Debugf("get content list ,total %d record", total)

}

//get one conetent by id
func getContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	i := q.Get("id")
	t := q.Get("type")
	status := q.Get("status") //get content by status
	logger.Debugf("Get content %s by id :%s , status %s", t, i, status)
	contentType, ok := item.Types[t]
	if !ok {
		//fmt.Fprintf(res, item.ErrTypeNotRegistered.Error(), t)
		logger.Error("Input data error ")
		renderJSON(w, r, ReturnData{
			RetCode: -5,
			Msg:     item.ErrTypeNotRegistered.Error(),
		})
		return
	}

	post := contentType()

	if i != "" {
		if status == "pending" {
			t = t + "__pending"
		}

		data, err := db.Content(t + ":" + i)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		if len(data) < 1 || data == nil {
			logger.Error("Content is not exists")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		retdata := map[string]interface{}{}
		//err = json.Unmarshal(data, post)
		err = json.Unmarshal(data, &retdata)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//fmt.Println(retdata)
		hook, ok := post.(item.Hookable)

		if ok {
			// hook before response
			fields, hasSubContent := hook.EnableSubContent()
			if hasSubContent {

				logger.Debug("Now process sub-content ")

				for index := range fields {
					fieldname := fields[index]
					data, err := db.GetSubContent(t+":"+i, fieldname)
					//fmt.Println(t + specifier + ":" + fmt.Sprint(retData[kk]["id"]))
					if err == nil {
						outdata := []map[string]interface{}{}
						err := json.Unmarshal(data, &outdata)
						if err == nil {
							retdata[fieldname] = outdata
						}
					}

				}

			}

		}

		renderJSON(w, r, map[string]interface{}{
			"retCode": 0,
			"msg":     "ok",
			"data":    retdata,
		})
		return
	}

	renderJSON(w, r, ReturnData{
		RetCode: -3,
		Msg:     "Not Found",
	})

}

//to update or create content
func updateContent(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		q := r.URL.Query()
		t := q.Get("type") //内容 类型
		logger.Debugf("To update content %s,from %s", t, GetIP(r))
		updateData := getJsonFromBody(r) // get the update content
		cid := q.Get("id")               // get update content id
		if updateData == nil || cid == "" {
			renderJSON(w, r, ReturnData{
				RetCode: -1,
				Msg:     "No Input Data",
			})
			return
		}

		ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))
		//up := r.PostForm.Set("updated", ts)
		updateData["updated"] = ts

		pt := t
		if strings.Contains(t, "__") {
			pt = strings.Split(t, "__")[0]
		}

		p, ok := item.Types[pt]
		if !ok {
			logger.Debugf("Type", t, "is not a content type. Cannot edit or save.")
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		post := p()
		hook, ok := post.(item.Hookable) //execute hook program
		if !ok {
			logger.Debug("Type", pt, "does not implement item.Hookable or embed item.Item.")
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		err = hook.BeforeAdminUpdate(w, r)
		if err != nil {
			logger.Debug("Error running BeforeAdminUpdate method in editHandler for:", t, err)
			return
		}

		err = hook.BeforeSave(w, r) ///before save
		if err != nil {
			logger.Debug("Error running BeforeSave method in editHandler for:", t, err)
			return
		}

		upp := formatData(updateData)
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		dec.SetAliasTag("json")
		err = dec.Decode(post, upp)
		if err != nil {
			logger.Debug("Error decoding post form for edit handler:", t, err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		id, err := db.UpdateContent(t+":"+cid, upp)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// set the target in the context so user can get saved value from db in hook
		ctx := context.WithValue(r.Context(), "target", fmt.Sprintf("%s:%d", t, id))

		r = r.WithContext(ctx)

		err = hook.AfterSave(w, r) //invoke the after save
		if err != nil {
			logger.Error("Error running AfterSave method in editHandler for:", t, err)
			return
		}

		err = hook.AfterAdminUpdate(w, r) //invoker admin update hoook
		if err != nil {
			logger.Error("Error running AfterAdminUpdate method in editHandler for:", t, err)
			return
		}

		renderJSON(w, r, ReturnData{
			RetCode: 0,
			Msg:     "Done",
		})

		///http.Redirect(res, req, redir, http.StatusFound)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//to update or create content
func createContent(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		q := r.URL.Query()
		t := q.Get("type") //内容 类型
		logger.Debugf("To create/update content %s,from %s", t, GetIP(r))
		updateData := getJsonFromBody(r) // get the update content
		//cid := q.Get("id")               // get update content id
		if updateData == nil {
			renderJSON(w, r, ReturnData{
				RetCode: -1,
				Msg:     "No Input Data",
			})
			return
		}
		/* if cid == "" {
			cid = "-1" // 没有id就是create
			logger.Debug("It is create request")
		} else {
			logger.Debug("It is update request")
		} */
		ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))
		//up := r.PostForm.Set("updated", ts)
		updateData["updated"] = ts

		pt := t
		if strings.Contains(t, "__") {
			pt = strings.Split(t, "__")[0]
		}

		p, ok := item.Types[pt]
		if !ok {
			logger.Debugf("Type", t, "is not a content type. Cannot edit or save.")
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		post := p()
		hook, ok := post.(item.Hookable) //execute hook program
		if !ok {
			logger.Debug("Type", pt, "does not implement item.Hookable or embed item.Item.")
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		// Let's be nice and make a proper item for the Hookable methods
		/* 	dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		dec.SetAliasTag("json")
		err = dec.Decode(post, r.PostForm)
		if err != nil {
			logger.Debug("Error decoding post form for edit handler:", t, err)
			w.WriteHeader(http.StatusBadRequest)

			return
		} */

		err = hook.BeforeAdminCreate(w, r)
		if err != nil {
			logger.Debug("Error running BeforeAdminCreate method in editHandler for:", t, err)
			return
		}

		err = hook.BeforeSave(w, r) ///before save
		if err != nil {
			logger.Debug("Error running BeforeSave method in editHandler for:", t, err)
			return
		}

		upp := formatData(updateData)

		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		dec.SetAliasTag("json")
		err = dec.Decode(post, upp)
		if err != nil {
			logger.Debug("Error decoding post form for edit handler:", t, err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		id, err := db.SetContent(t+":-1", upp)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		// set the target in the context so user can get saved value from db in hook
		ctx := context.WithValue(r.Context(), "target", fmt.Sprintf("%s:%d", t, id))

		r = r.WithContext(ctx)

		err = hook.AfterSave(w, r) //invoke the after save
		if err != nil {
			logger.Error("Error running AfterSave method in editHandler for:", t, err)
			return
		}

		err = hook.AfterAdminCreate(w, r) //invoker admin create hoook
		if err != nil {
			logger.Error("Error running AfterAdminUpdate method in editHandler for:", t, err)
			return
		}

		renderJSON(w, r, ReturnData{
			RetCode: 0,
			Msg:     "Done",
		})

		///http.Redirect(res, req, redir, http.StatusFound)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// to search content get back a search result
func searchContent(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	t := q.Get("type")
	search := q.Get("q")
	logger.Debugf("Search content %s with %s from %s", t, search, GetIP(r))
	status := q.Get("status")
	regexsearch := q.Get("r")
	var specifier string

	if t == "" || (search == "" && regexsearch == "") {
		logger.Debugf("Search parameter missing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if status == "pending" {
		specifier = "__" + status
	}

	posts := db.ContentAll(t + specifier)
	//b := &bytes.Buffer{}
	//pt, ok := item.Types[t]
	/* 	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		logger.Debug("No such content type ")
		return
	} */

	//post := pt()
	retData := make([]map[string]interface{}, 0)
	match := strings.ToLower(search)

	for i := range posts {
		// skip posts that don't have any matching search criteria
		if search != "" { // contain str
			all := strings.ToLower(string(posts[i]))

			if !strings.Contains(all, match) {
				continue
			}
			item := make(map[string]interface{})
			err := json.Unmarshal(posts[i], &item)

			if err != nil {
				logger.Debug("Error unmarshal search result json into", t, err, posts[i])
				continue
			}
			retData = append(retData, item)
		} else if regexsearch != "" { // use regex to search
			re := regexp.MustCompile(regexsearch)
			if re.Match(posts[i]) {
				item := make(map[string]interface{})
				err := json.Unmarshal(posts[i], &item)

				if err != nil {
					logger.Debug("Error unmarshal search result json into", t, err, posts[i])
					continue
				}
				//fmt.Println(item)
				retData = append(retData, item)
			}
		}
	}
	total := len(posts)
	meta := MetaData{
		Total:     uint(total),
		PageCount: 1,
		Page:      0,
		Order:     "",
		PageSize:  len(retData), //-1 means all
	}
	returnStructData(w, r, retData, meta)

}

func contentsRestHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	order := strings.ToLower(q.Get("order"))
	if order != "asc" {
		order = "desc"
	}

	status := q.Get("status")

	if _, ok := item.Types[t]; !ok {
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	pt := item.Types[t]()

	p, ok := pt.(editor.Editable)
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	var hasExt bool
	_, ok = pt.(api.Createable)
	if ok {
		hasExt = true
	}

	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if q.Get("count") == "" {
			count = 10
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}
	}

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}
	}

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}

	var specifier string
	if status == "public" || status == "" {
		specifier = "__sorted"
	} else if status == "pending" {
		specifier = "__pending"
	}

	b := &bytes.Buffer{}
	var total int
	var posts [][]byte

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="col s8">
						<div class="row">
							<div class="card-title col s7">` + t + ` Items</div>
							<div class="col s5 input-field inline">
								<select class="browser-default __dms sort-order">
									<option value="DESC">New to Old</option>
									<option value="ASC">Old to New</option>
								</select>
								<label class="active">Sort:</label>
							</div>	
							<script>
								$(function() {
									var sort = $('select.__dms.sort-order');

									sort.on('change', function() {
										var path = window.location.pathname;
										var s = sort.val();
										var t = getParam('type');
										var status = getParam('status');

										if (status == "") {
											status = "public";
										}

										window.location.replace(path + '?type=' + t + '&order=' + s + '&status=' + status);
									});

									var order = getParam('order');
									if (order !== '') {
										sort.val(order);
									}
									
								});
							</script>
						</div>
					</div>
					<form class="col s4" action="/admin/contents/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all ` + t + ` fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
							<input type="hidden" name="status" value="` + status + `" />
						</div>
                    </form>	
					</div>`
	if hasExt {
		if status == "" {
			q.Set("status", "public")
		}

		// always start from top of results when changing public/pending
		q.Del("count")
		q.Del("offset")

		q.Set("status", "public")
		publicURL := req.URL.Path + "?" + q.Encode()

		q.Set("status", "pending")
		pendingURL := req.URL.Path + "?" + q.Encode()

		switch status {
		case "public", "":
			// get __sorted posts of type t from the db
			total, posts = db.Query(t+specifier, opts)

			html += `<div class="row externalable">
					<span class="description">Status:</span> 
					<span class="active">Public</span>
					&nbsp;&vert;&nbsp;
					<a href="` + pendingURL + `">Pending</a>
				</div>`

			for i := range posts {
				err := json.Unmarshal(posts[i], &p)
				if err != nil {
					log.Println("Error unmarshal json into", t, err, string(posts[i]))

					post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
					_, err := b.Write([]byte(post))
					if err != nil {
						log.Println(err)

						res.WriteHeader(http.StatusInternalServerError)
						errView, err := Error500()
						if err != nil {
							log.Println(err)
						}

						res.Write(errView)
						return
					}

					continue
				}

				post := adminPostListItem(p, t, status)
				_, err = b.Write(post)
				if err != nil {
					log.Println(err)

					res.WriteHeader(http.StatusInternalServerError)
					errView, err := Error500()
					if err != nil {
						log.Println(err)
					}

					res.Write(errView)
					return
				}
			}

		case "pending":
			// get __pending posts of type t from the db
			total, posts = db.Query(t+"__pending", opts)

			html += `<div class="row externalable">
					<span class="description">Status:</span> 
					<a href="` + publicURL + `">Public</a>
					&nbsp;&vert;&nbsp;
					<span class="active">Pending</span>					
				</div>`

			for i := len(posts) - 1; i >= 0; i-- {
				err := json.Unmarshal(posts[i], &p)
				if err != nil {
					log.Println("Error unmarshal json into", t, err, string(posts[i]))

					post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
					_, err := b.Write([]byte(post))
					if err != nil {
						log.Println(err)

						res.WriteHeader(http.StatusInternalServerError)
						errView, err := Error500()
						if err != nil {
							log.Println(err)
						}

						res.Write(errView)
						return
					}
					continue
				}

				post := adminPostListItem(p, t, status)
				_, err = b.Write(post)
				if err != nil {
					log.Println(err)

					res.WriteHeader(http.StatusInternalServerError)
					errView, err := Error500()
					if err != nil {
						log.Println(err)
					}

					res.Write(errView)
					return
				}
			}
		}

	} else {
		total, posts = db.Query(t+specifier, opts)

		for i := range posts {
			err := json.Unmarshal(posts[i], &p)
			if err != nil {
				log.Println("Error unmarshal json into", t, err, string(posts[i]))

				post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
				_, err := b.Write([]byte(post))
				if err != nil {
					log.Println(err)

					res.WriteHeader(http.StatusInternalServerError)
					errView, err := Error500()
					if err != nil {
						log.Println(err)
					}

					res.Write(errView)
					return
				}
				continue
			}

			post := adminPostListItem(p, t, status)
			_, err = b.Write(post)
			if err != nil {
				log.Println(err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					log.Println(err)
				}

				res.Write(errView)
				return
			}
		}
	}

	html += `<ul class="posts row">`

	_, err = b.Write([]byte(`</ul>`))
	if err != nil {
		log.Println(err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			log.Println(err)
		}

		res.Write(errView)
		return
	}

	statusDisabled := "disabled"
	prevStatus := ""
	nextStatus := ""
	// total may be less than 10 (default count), so reset count to match total
	if total < count {
		count = total
	}
	// nothing previous to current list
	if offset == 0 {
		prevStatus = statusDisabled
	}
	// nothing after current list
	if (offset+1)*count >= total {
		nextStatus = statusDisabled
	}

	// set up pagination values
	urlFmt := req.URL.Path + "?count=%d&offset=%d&&order=%s&status=%s&type=%s"
	prevURL := fmt.Sprintf(urlFmt, count, offset-1, order, status, t)
	nextURL := fmt.Sprintf(urlFmt, count, offset+1, order, status, t)
	start := 1 + count*offset
	end := start + count - 1

	if total < end {
		end = total
	}

	pagination := fmt.Sprintf(`
	<ul class="pagination row">
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_left</i></a></li>
		<li class="col s8">%d to %d of %d</li>
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_right</i></a></li>
	</ul>
	`, prevStatus, prevURL, start, end, total, nextStatus, nextURL)

	// show indicator that a collection of items will be listed implicitly, but
	// that none are created yet
	if total < 1 {
		pagination = `
		<ul class="pagination row">
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_left</i></a></li>
			<li class="col s8">0 to 0 of 0</li>
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_right</i></a></li>
		</ul>
		`
	}

	_, err = b.Write([]byte(pagination + `</div></div>`))
	if err != nil {
		log.Println(err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			log.Println(err)
		}

		res.Write(errView)
		return
	}

	script := `
	<script>
		$(function() {
			var del = $('.quick-delete-post.__dms span');
			del.on('click', function(e) {
				if (confirm("[Dms] Please confirm:\n\nAre you sure you want to delete this post?\nThis cannot be undone.")) {
					$(e.target).parent().submit();
				}
			});
		});

		// disable link from being clicked if parent is 'disabled'
		$(function() {
			$('ul.pagination li.disabled a').on('click', function(e) {
				e.preventDefault();
			});
		});
	</script>
	`

	btn := `<div class="col s3">
		<a href="/admin/edit?type=` + t + `" class="btn new-post waves-effect waves-light">
			New ` + t + `
		</a>`

	if _, ok := pt.(format.CSVFormattable); ok {
		btn += `<br/>
				<a href="/admin/contents/export?type=` + t + `&format=csv" class="green darken-4 btn export-post waves-effect waves-light">
					<i class="material-icons left">system_update_alt</i>
					CSV
				</a>`
	}

	html += b.String() + script + btn + `</div></div>`

	adminView, err := Admin([]byte(html))
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}

// to approve a content
func approveContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	t := q.Get("type") //内容 类型
	logger.Debugf("To approve content %s,from %s", t, GetIP(r))
	content := getJsonFromBody(r)
	if content == nil {
		logger.Error("No Content")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pendingID := q.Get("id")

	if strings.Contains(t, "__") { // look like this tricky
		t = strings.Split(t, "__")[0]
	}

	post := item.Types[t]()

	// run hooks
	hook, ok := post.(item.Hookable) //run hookevent
	if !ok {
		logger.Error("Type", t, "does not implement item.Hookable or embed item.Item.")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// check if we have a Mergeable
	m, ok := post.(editor.Mergeable) // run merageble hooke
	if !ok {
		logger.Error("Content type", t, "must implement editor.Mergeable before it can be approved.")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	/* 	dec := schema.NewDecoder()
	   	dec.IgnoreUnknownKeys(true)
	   	dec.SetAliasTag("json")
	   	err = dec.Decode(post, req.Form)
	   	if err != nil {
	   		log.Println("Error decoding post form for content approval:", t, err)
	   		res.WriteHeader(http.StatusInternalServerError)
	   		errView, err := Error500()
	   		if err != nil {
	   			return
	   		}

	   		res.Write(errView)
	   		return
	   	}
	*/
	err = hook.BeforeApprove(w, r) // run beforeapprove hook program
	if err != nil {
		logger.Warn("Error running BeforeApprove hook in approveContentHandler for:", t, err)
		return
	}

	// call its Approve method
	err = m.Approve(w, r) // to approve event
	if err != nil {
		logger.Warn("Error running Approve method in approveContentHandler for:", t, err)
		return
	}

	err = hook.AfterApprove(w, r) // run afterapprove hook
	if err != nil {
		logger.Warn("Error running AfterApprove hook in approveContentHandler for:", t, err)
		return
	}

	err = hook.BeforeSave(w, r) // run beforesave hook
	if err != nil {
		logger.Warn("Error running BeforeSave hook in approveContentHandler for:", t, err)
		return
	}

	newentry := formatData(content)
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, newentry)
	if err != nil {
		logger.Debug("Error decoding post form for content approval:", t, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Store the content in the bucket t
	//fmt.Printf("%v", newentry)
	id, err := db.SetContent(t+":-1", newentry)
	if err != nil {
		logger.Warn("Error storing content in approveContentHandler for:", t, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// set the target in the context so user can get saved value from db in hook
	ctx := context.WithValue(r.Context(), "target", fmt.Sprintf("%s:%d", t, id))
	r = r.WithContext(ctx)

	err = hook.AfterSave(w, r)
	if err != nil {
		logger.Warn("Error running AfterSave hook in approveContentHandler for:", t, err)
		return
	}

	pendingBucks := t + PENDINGSuffix
	if pendingID != "" {
		err = db.DeleteContent(pendingBucks + ":" + pendingID)
		if err != nil {
			logger.Warn("Failed to remove content after approval:", err)
		}
	}
	// redirect to the new approved content's editor

	renderJSON(w, r, ReturnData{
		RetCode: 0,
		Msg:     "Approved with no error",
	})
	//http.Redirect(res, req, redir, http.StatusFound)
}

// to approve a content
func rejectContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	t := q.Get("type") //内容 类型
	id := q.Get("id")
	logger.Debugf("To reject content %s,from %s", t, GetIP(r))
	if id == "" {
		logger.Error("No Content ID input ")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ct := t

	if t == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// catch specifier suffix from delete form value
	if strings.Contains(t, "__") {
		spec := strings.Split(t, "__")
		ct = spec[0]
	}

	p, ok := item.Types[ct]
	if !ok {
		logger.Error("Type", t, "does not implement item.Hookable or embed item.Item.")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	post := p()
	hook, ok := post.(item.Hookable)
	if !ok {
		logger.Error("Type", t, "does not implement item.Hookable or embed item.Item.")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	_, err := db.Content(t + PENDINGSuffix + ":" + id) //target is pending bucket
	if err != nil {
		logger.Error("no reject content ", t+":"+id, err)
		renderJSON(w, r, ReturnData{
			RetCode: -8,
			Msg:     "No such content",
		})
		return
	}
	/*
		err = json.Unmarshal(data, post)
		if err != nil {
			logger.Error("Error unmarshalling ", t, "=", id, err, " Hooks will be called on a zero-value.")
		}
	*/
	// call reject hook
	err = hook.BeforeReject(w, r)
	if err != nil {
		logger.Error("Error running BeforeReject method in deleteHandler for:", t, err)
		return
	}

	err = hook.BeforeAdminDelete(w, r)
	if err != nil {
		logger.Error("Error running BeforeAdminDelete method in deleteHandler for:", t, err)
		renderJSON(w, r, ReturnData{
			RetCode: -8,
			Msg:     "failed on before admin delete check",
		})
		return
	}

	err = hook.BeforeDelete(w, r)
	if err != nil {
		logger.Error("Error running BeforeDelete method in deleteHandler for:", t, err)
		renderJSON(w, r, ReturnData{
			RetCode: -8,
			Msg:     "failed on before delete check",
		})
		return
	}
	logger.Debug("Now to delete the content ", t, ": ", id)
	err = db.DeleteContent(t + PENDINGSuffix + ":" + id)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = hook.AfterDelete(w, r)
	if err != nil {
		logger.Error("Error running AfterDelete method in deleteHandler for:", t, err)
		return
	}

	err = hook.AfterAdminDelete(w, r)
	if err != nil {
		logger.Error("Error running AfterDelete method in deleteHandler for:", t, err)
		return
	}
	// call reject after hook
	err = hook.AfterReject(w, r)
	if err != nil {
		logger.Error("Error running AfterReject method in deleteHandler for:", t, err)
		return
	}

	renderJSON(w, r, ReturnData{
		RetCode: 0,
		Msg:     "Done",
	})

	//http.Redirect(res, req, redir, http.StatusFound)
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	t := q.Get("type") //内容 类型
	id := q.Get("id")
	logger.Debugf("To delete  content %s,from %s", t, GetIP(r))
	if id == "" {
		logger.Error("No Content ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ct := t

	if t == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// catch specifier suffix from delete form value
	if strings.Contains(t, "__") {
		spec := strings.Split(t, "__")
		ct = spec[0]
	}

	p, ok := item.Types[ct]
	if !ok {
		logger.Error("Type", t, "does not implement item.Hookable or embed item.Item.")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	post := p()
	hook, ok := post.(item.Hookable)
	if !ok {
		logger.Error("Type", t, "does not implement item.Hookable or embed item.Item.")
		w.WriteHeader(http.StatusBadRequest)

		return
	}
	/*
		data, err := db.Content(t + ":" + id)
		if err != nil {
			logger.Error("Error in db.Content ", t+":"+id, err)
			renderJSON(w, r, ReturnData{
				RetCode: -8,
				Msg:     "No such content",
			})
			return
		} */
	/*
		err = json.Unmarshal(data, post)
		if err != nil {
			logger.Error("Error unmarshalling ", t, "=", id, err, " Hooks will be called on a zero-value.")
		}
	*/
	reject := r.URL.Query().Get("reject")
	if reject == "true" {
		err = hook.BeforeReject(w, r)
		if err != nil {
			logger.Error("Error running BeforeReject method in deleteHandler for:", t, err)
			return
		}
	}

	err = hook.BeforeAdminDelete(w, r)
	if err != nil {
		logger.Error("Error running BeforeAdminDelete method in deleteHandler for:", t, err)
		renderJSON(w, r, ReturnData{
			RetCode: -8,
			Msg:     "failed on before admin delete check",
		})
		return
	}

	err = hook.BeforeDelete(w, r)
	if err != nil {
		logger.Error("Error running BeforeDelete method in deleteHandler for:", t, err)
		renderJSON(w, r, ReturnData{
			RetCode: -8,
			Msg:     "failed on before delete check",
		})
		return
	}
	logger.Debug("Now to delete the content ", t, ": ", id)
	err = db.DeleteContent(t + ":" + id)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = hook.AfterDelete(w, r)
	if err != nil {
		logger.Error("Error running AfterDelete method in deleteHandler for:", t, err)
		return
	}

	err = hook.AfterAdminDelete(w, r)
	if err != nil {
		logger.Error("Error running AfterDelete method in deleteHandler for:", t, err)
		return
	}

	if reject == "true" {
		err = hook.AfterReject(w, r)
		if err != nil {
			logger.Error("Error running AfterReject method in deleteHandler for:", t, err)
			return
		}
	}
	renderJSON(w, r, ReturnData{
		RetCode: 0,
		Msg:     "Done",
	})

}

// to upload a file to server media library
func uploadMediaContent(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Upload the media content from :", GetIP(r))

	urlPaths, err := upload.StoreFiles(r)
	if err != nil {
		logger.Error("Couldn't store file uploads.", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//fmt.Println(urlPaths)
	renderJSON(w, r, map[string]interface{}{
		"retCode": 0,
		"msg":     "ok",
		"data":    urlPaths,
	})
	//	r.Header().Set("Content-Type", "application/json")
	//	w.Write([]byte(`{"data": [{"url": "` + urlPaths["file"] + `"}]}`))

}

//delete the upload file
func deleteMediaContent(w http.ResponseWriter, r *http.Request) {

	logger.Debugf("delete upload file  from ip:", GetIP(r))
	q := r.URL.Query()
	id := q.Get("id")
	t := db.DB__uploads

	if id == "" || t == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	post := interface{}(&item.FileUpload{})
	hook, ok := post.(item.Hookable)
	if !ok {
		logger.Debug("Type", t, "does not implement item.Hookable or embed item.Item.")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	err = hook.BeforeDelete(w, r)
	if err != nil {
		logger.Error("Error running BeforeDelete method in deleteHandler for:", t, err)

		return
	}

	dbTarget := t + ":" + id

	// delete from file system, if good, we continue to delete
	// from database, if bad error 500
	err = deleteUploadFromDisk(dbTarget)
	if err != nil {
		logger.Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.DeleteUpload(dbTarget)
	if err != nil {
		logger.Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = hook.AfterDelete(w, r)
	if err != nil {
		logger.Debug("Error running AfterDelete method in deleteHandler for:", t, err)
		return
	}

	renderJSON(w, r, map[string]interface{}{
		"retCode": 0,
		"msg":     "Done",
	})
}

func approveContentRestHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		errView, err := Error405()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	pendingID := req.FormValue("id")

	t := req.FormValue("type")
	if strings.Contains(t, "__") {
		t = strings.Split(t, "__")[0]
	}

	post := item.Types[t]()

	// run hooks
	hook, ok := post.(item.Hookable)
	if !ok {
		log.Println("Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	// check if we have a Mergeable
	m, ok := post.(editor.Mergeable)
	if !ok {
		log.Println("Content type", t, "must implement editor.Mergeable before it can be approved.")
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, req.Form)
	if err != nil {
		log.Println("Error decoding post form for content approval:", t, err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	err = hook.BeforeApprove(res, req)
	if err != nil {
		log.Println("Error running BeforeApprove hook in approveContentHandler for:", t, err)
		return
	}

	// call its Approve method
	err = m.Approve(res, req)
	if err != nil {
		log.Println("Error running Approve method in approveContentHandler for:", t, err)
		return
	}

	err = hook.AfterApprove(res, req)
	if err != nil {
		log.Println("Error running AfterApprove hook in approveContentHandler for:", t, err)
		return
	}

	err = hook.BeforeSave(res, req)
	if err != nil {
		log.Println("Error running BeforeSave hook in approveContentHandler for:", t, err)
		return
	}

	// Store the content in the bucket t
	id, err := db.SetContent(t+":-1", req.Form)
	if err != nil {
		log.Println("Error storing content in approveContentHandler for:", t, err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	// set the target in the context so user can get saved value from db in hook
	ctx := context.WithValue(req.Context(), "target", fmt.Sprintf("%s:%d", t, id))
	req = req.WithContext(ctx)

	err = hook.AfterSave(res, req)
	if err != nil {
		log.Println("Error running AfterSave hook in approveContentHandler for:", t, err)
		return
	}

	if pendingID != "" {
		err = db.DeleteContent(req.FormValue("type") + ":" + pendingID)
		if err != nil {
			log.Println("Failed to remove content after approval:", err)
		}
	}

	// redirect to the new approved content's editor
	redir := req.URL.Scheme + req.URL.Host + strings.TrimSuffix(req.URL.Path, "/approve")
	redir += fmt.Sprintf("?type=%s&id=%d", t, id)
	http.Redirect(res, req, redir, http.StatusFound)
}

func editRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		q := req.URL.Query()
		i := q.Get("id")
		t := q.Get("type")
		status := q.Get("status")

		contentType, ok := item.Types[t]
		if !ok {
			fmt.Fprintf(res, item.ErrTypeNotRegistered.Error(), t)
			return
		}
		post := contentType()

		if i != "" {
			if status == "pending" {
				t = t + "__pending"
			}

			data, err := db.Content(t + ":" + i)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			if len(data) < 1 || data == nil {
				res.WriteHeader(http.StatusNotFound)
				errView, err := Error404()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			err = json.Unmarshal(data, post)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}
		} else {
			item, ok := post.(item.Identifiable)
			if !ok {
				log.Println("Content type", t, "doesn't implement item.Identifiable")
				return
			}

			item.SetItemID(-1)
		}

		m, err := manager.Manage(post.(editor.Editable), t)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		adminView, err := Admin(m)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		cid := req.FormValue("id")
		t := req.FormValue("type")
		ts := req.FormValue("timestamp")
		up := req.FormValue("updated")

		// create a timestamp if one was not set
		if ts == "" {
			ts = fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))
			req.PostForm.Set("timestamp", ts)
		}

		if up == "" {
			req.PostForm.Set("updated", ts)
		}

		urlPaths, err := upload.StoreFiles(req)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		for name, urlPath := range urlPaths {
			req.PostForm.Set(name, urlPath)
		}

		// check for any multi-value fields (ex. checkbox fields)
		// and correctly format for db storage. Essentially, we need
		// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
		fieldOrderValue := make(map[string]map[string][]string)
		for k, v := range req.PostForm {
			if strings.Contains(k, ".") {
				fo := strings.Split(k, ".")

				// put the order and the field value into map
				field := string(fo[0])
				order := string(fo[1])
				if len(fieldOrderValue[field]) == 0 {
					fieldOrderValue[field] = make(map[string][]string)
				}

				// orderValue is 0:[?type=Thing&id=1]
				orderValue := fieldOrderValue[field]
				orderValue[order] = v
				fieldOrderValue[field] = orderValue

				// discard the post form value with name.N
				req.PostForm.Del(k)
			}

		}

		// add/set the key & value to the post form in order
		for f, ov := range fieldOrderValue {
			for i := 0; i < len(ov); i++ {
				position := fmt.Sprintf("%d", i)
				fieldValue := ov[position]

				if req.PostForm.Get(f) == "" {
					for i, fv := range fieldValue {
						if i == 0 {
							req.PostForm.Set(f, fv)
						} else {
							req.PostForm.Add(f, fv)
						}
					}
				} else {
					for _, fv := range fieldValue {
						req.PostForm.Add(f, fv)
					}
				}
			}
		}

		pt := t
		if strings.Contains(t, "__") {
			pt = strings.Split(t, "__")[0]
		}

		p, ok := item.Types[pt]
		if !ok {
			log.Println("Type", t, "is not a content type. Cannot edit or save.")
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		post := p()
		hook, ok := post.(item.Hookable)
		if !ok {
			log.Println("Type", pt, "does not implement item.Hookable or embed item.Item.")
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// Let's be nice and make a proper item for the Hookable methods
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		dec.SetAliasTag("json")
		err = dec.Decode(post, req.PostForm)
		if err != nil {
			log.Println("Error decoding post form for edit handler:", t, err)
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		if cid == "-1" {
			err = hook.BeforeAdminCreate(res, req)
			if err != nil {
				log.Println("Error running BeforeAdminCreate method in editHandler for:", t, err)
				return
			}
		} else {
			err = hook.BeforeAdminUpdate(res, req)
			if err != nil {
				log.Println("Error running BeforeAdminUpdate method in editHandler for:", t, err)
				return
			}
		}

		err = hook.BeforeSave(res, req)
		if err != nil {
			log.Println("Error running BeforeSave method in editHandler for:", t, err)
			return
		}

		id, err := db.SetContent(t+":"+cid, req.PostForm)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// set the target in the context so user can get saved value from db in hook
		ctx := context.WithValue(req.Context(), "target", fmt.Sprintf("%s:%d", t, id))
		req = req.WithContext(ctx)

		err = hook.AfterSave(res, req)
		if err != nil {
			log.Println("Error running AfterSave method in editHandler for:", t, err)
			return
		}

		if cid == "-1" {
			err = hook.AfterAdminCreate(res, req)
			if err != nil {
				log.Println("Error running AfterAdminUpdate method in editHandler for:", t, err)
				return
			}
		} else {
			err = hook.AfterAdminUpdate(res, req)
			if err != nil {
				log.Println("Error running AfterAdminUpdate method in editHandler for:", t, err)
				return
			}
		}

		scheme := req.URL.Scheme
		host := req.URL.Host
		path := req.URL.Path
		sid := fmt.Sprintf("%d", id)
		redir := scheme + host + path + "?type=" + pt + "&id=" + sid

		if req.URL.Query().Get("status") == "pending" {
			redir += "&status=pending"
		}

		http.Redirect(res, req, redir, http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func deleteRestHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	id := req.FormValue("id")
	t := req.FormValue("type")
	ct := t

	if id == "" || t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// catch specifier suffix from delete form value
	if strings.Contains(t, "__") {
		spec := strings.Split(t, "__")
		ct = spec[0]
	}

	p, ok := item.Types[ct]
	if !ok {
		log.Println("Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	post := p()
	hook, ok := post.(item.Hookable)
	if !ok {
		log.Println("Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	data, err := db.Content(t + ":" + id)
	if err != nil {
		log.Println("Error in db.Content ", t+":"+id, err)
		return
	}

	err = json.Unmarshal(data, post)
	if err != nil {
		log.Println("Error unmarshalling ", t, "=", id, err, " Hooks will be called on a zero-value.")
	}

	reject := req.URL.Query().Get("reject")
	if reject == "true" {
		err = hook.BeforeReject(res, req)
		if err != nil {
			log.Println("Error running BeforeReject method in deleteHandler for:", t, err)
			return
		}
	}

	err = hook.BeforeAdminDelete(res, req)
	if err != nil {
		log.Println("Error running BeforeAdminDelete method in deleteHandler for:", t, err)
		return
	}

	err = hook.BeforeDelete(res, req)
	if err != nil {
		log.Println("Error running BeforeDelete method in deleteHandler for:", t, err)
		return
	}

	err = db.DeleteContent(t + ":" + id)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = hook.AfterDelete(res, req)
	if err != nil {
		log.Println("Error running AfterDelete method in deleteHandler for:", t, err)
		return
	}

	err = hook.AfterAdminDelete(res, req)
	if err != nil {
		log.Println("Error running AfterDelete method in deleteHandler for:", t, err)
		return
	}

	if reject == "true" {
		err = hook.AfterReject(res, req)
		if err != nil {
			log.Println("Error running AfterReject method in deleteHandler for:", t, err)
			return
		}
	}

	redir := strings.TrimSuffix(req.URL.Scheme+req.URL.Host+req.URL.Path, "/edit/delete")
	redir = redir + "/contents?type=" + ct
	http.Redirect(res, req, redir, http.StatusFound)
}

func deleteUploadRestHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	logger.Debugf("delete request is %v", req)

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	id := req.FormValue("id")
	t := db.DB__uploads

	if id == "" || t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	post := interface{}(&item.FileUpload{})
	hook, ok := post.(item.Hookable)
	if !ok {
		log.Println("Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	err = hook.BeforeDelete(res, req)
	if err != nil {
		log.Println("Error running BeforeDelete method in deleteHandler for:", t, err)
		return
	}

	dbTarget := t + ":" + id

	// delete from file system, if good, we continue to delete
	// from database, if bad error 500
	err = deleteUploadFromDisk(dbTarget)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.DeleteUpload(dbTarget)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = hook.AfterDelete(res, req)
	if err != nil {
		log.Println("Error running AfterDelete method in deleteHandler for:", t, err)
		return
	}

	redir := "/admin/uploads"
	http.Redirect(res, req, redir, http.StatusFound)
}

func editUploadRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		q := req.URL.Query()
		i := q.Get("id")
		t := db.DB__uploads

		post := &item.FileUpload{}

		if i != "" {
			data, err := db.Upload(t + ":" + i)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			if len(data) < 1 || data == nil {
				res.WriteHeader(http.StatusNotFound)
				errView, err := Error404()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			err = json.Unmarshal(data, post)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}
		} else {
			it, ok := interface{}(post).(item.Identifiable)
			if !ok {
				log.Println("Content type", t, "doesn't implement item.Identifiable")
				return
			}

			it.SetItemID(-1)
		}

		m, err := manager.Manage(interface{}(post).(editor.Editable), t)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		adminView, err := Admin(m)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		t := req.FormValue("type")
		pt := db.DB__uploads
		ts := req.FormValue("timestamp")
		up := req.FormValue("updated")

		// create a timestamp if one was not set
		if ts == "" {
			ts = fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))
			req.PostForm.Set("timestamp", ts)
		}

		if up == "" {
			req.PostForm.Set("updated", ts)
		}

		post := interface{}(&item.FileUpload{})
		hook, ok := post.(item.Hookable)
		if !ok {
			log.Println("Type", pt, "does not implement item.Hookable or embed item.Item.")
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		err = hook.BeforeSave(res, req)
		if err != nil {
			log.Println("Error running BeforeSave method in editHandler for:", t, err)
			return
		}

		// StoreFiles has the SetUpload call (which is equivalent of SetContent in other handlers)
		urlPaths, err := upload.StoreFiles(req)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		for name, urlPath := range urlPaths {
			req.PostForm.Set(name, urlPath)
		}

		// check for any multi-value fields (ex. checkbox fields)
		// and correctly format for db storage. Essentially, we need
		// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
		fieldOrderValue := make(map[string]map[string][]string)
		ordVal := make(map[string][]string)
		for k, v := range req.PostForm {
			if strings.Contains(k, ".") {
				fo := strings.Split(k, ".")

				// put the order and the field value into map
				field := string(fo[0])
				order := string(fo[1])
				fieldOrderValue[field] = ordVal

				// orderValue is 0:[?type=Thing&id=1]
				orderValue := fieldOrderValue[field]
				orderValue[order] = v
				fieldOrderValue[field] = orderValue

				// discard the post form value with name.N
				req.PostForm.Del(k)
			}

		}

		// add/set the key & value to the post form in order
		for f, ov := range fieldOrderValue {
			for i := 0; i < len(ov); i++ {
				position := fmt.Sprintf("%d", i)
				fieldValue := ov[position]

				if req.PostForm.Get(f) == "" {
					for i, fv := range fieldValue {
						if i == 0 {
							req.PostForm.Set(f, fv)
						} else {
							req.PostForm.Add(f, fv)
						}
					}
				} else {
					for _, fv := range fieldValue {
						req.PostForm.Add(f, fv)
					}
				}
			}
		}

		err = hook.AfterSave(res, req)
		if err != nil {
			log.Println("Error running AfterSave method in editHandler for:", t, err)
			return
		}

		scheme := req.URL.Scheme
		host := req.URL.Host
		redir := scheme + host + "/admin/uploads"
		http.Redirect(res, req, redir, http.StatusFound)

	case http.MethodPut:
		urlPaths, err := upload.StoreFiles(req)
		if err != nil {
			log.Println("Couldn't store file uploads.", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(`{"data": [{"url": "` + urlPaths["file"] + `"}]}`))
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

/*
func editUploadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	urlPaths, err := upload.StoreFiles(req)
	if err != nil {
		log.Println("Couldn't store file uploads.", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write([]byte(`{"data": [{"url": "` + urlPaths["file"] + `"}]}`))
} */

func searchRestHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	search := q.Get("q")
	status := q.Get("status")
	var specifier string

	if t == "" || search == "" {
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	if status == "pending" {
		specifier = "__" + status
	}

	posts := db.ContentAll(t + specifier)
	b := &bytes.Buffer{}
	pt, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	post := pt()

	p := post.(editor.Editable)

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="card-title col s7">` + t + ` Results</div>	
					<form class="col s4" action="/admin/contents/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all ` + t + ` fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
							<input type="hidden" name="status" value="` + status + `" />
						</div>
                    </form>	
					</div>
					<ul class="posts row">`

	for i := range posts {
		// skip posts that don't have any matching search criteria
		match := strings.ToLower(search)
		all := strings.ToLower(string(posts[i]))
		if !strings.Contains(all, match) {
			continue
		}

		err := json.Unmarshal(posts[i], &p)
		if err != nil {
			log.Println("Error unmarshal search result json into", t, err, posts[i])

			post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
			_, err = b.Write([]byte(post))
			if err != nil {
				log.Println(err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					log.Println(err)
				}

				res.Write(errView)
				return
			}
			continue
		}

		post := adminPostListItem(p, t, status)
		_, err = b.Write([]byte(post))
		if err != nil {
			log.Println(err)

			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				log.Println(err)
			}

			res.Write(errView)
			return
		}
	}

	_, err := b.WriteString(`</ul></div></div>`)
	if err != nil {
		log.Println(err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			log.Println(err)
		}

		res.Write(errView)
		return
	}

	script := `
	<script>
		$(function() {
			var del = $('.quick-delete-post.__ponzu span');
			del.on('click', function(e) {
				if (confirm("[Ponzu] Please confirm:\n\nAre you sure you want to delete this post?\nThis cannot be undone.")) {
					$(e.target).parent().submit();
				}
			});
		});

		// disable link from being clicked if parent is 'disabled'
		$(function() {
			$('ul.pagination li.disabled a').on('click', function(e) {
				e.preventDefault();
			});
		});
	</script>
	`

	btn := `<div class="col s3">
		<a href="/admin/edit?type=` + t + `" class="btn new-post waves-effect waves-light">
			New ` + t + `
		</a>`

	html += b.String() + script + btn + `</div></div>`

	adminView, err := Admin([]byte(html))
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}

func uploadSearchRestHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := db.DB__uploads
	search := q.Get("q")
	status := q.Get("status")

	if t == "" || search == "" {
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	posts := db.UploadAll()
	b := &bytes.Buffer{}
	p := interface{}(&item.FileUpload{}).(editor.Editable)

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="card-title col s7">Uploads Results</div>	
					<form class="col s4" action="/admin/uploads/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all Upload fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
						</div>
                    </form>	
					</div>
					<ul class="posts row">`

	for i := range posts {
		// skip posts that don't have any matching search criteria
		match := strings.ToLower(search)
		all := strings.ToLower(string(posts[i]))
		if !strings.Contains(all, match) {
			continue
		}

		err := json.Unmarshal(posts[i], &p)
		if err != nil {
			log.Println("Error unmarshal search result json into", t, err, posts[i])

			post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
			_, err = b.Write([]byte(post))
			if err != nil {
				log.Println(err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					log.Println(err)
				}

				res.Write(errView)
				return
			}
			continue
		}

		post := adminPostListItem(p, t, status)
		_, err = b.Write([]byte(post))
		if err != nil {
			log.Println(err)

			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				log.Println(err)
			}

			res.Write(errView)
			return
		}
	}

	_, err := b.WriteString(`</ul></div></div>`)
	if err != nil {
		log.Println(err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			log.Println(err)
		}

		res.Write(errView)
		return
	}

	btn := `<div class="col s3"><a href="/admin/edit/upload" class="btn new-post waves-effect waves-light">New Upload</a></div></div>`
	html = html + b.String() + btn

	adminView, err := Admin([]byte(html))
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}

func addonsRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		all := db.AddonAll()
		list := &bytes.Buffer{}

		for i := range all {
			v := adminAddonListItem(all[i])
			_, err := list.Write(v)
			if err != nil {
				log.Println("Error writing bytes to addon list view:", err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					log.Println(err)
					return
				}

				res.Write(errView)
				return
			}
		}

		html := &bytes.Buffer{}
		open := `<div class="col s9 card">		
				<div class="card-content">
				<div class="row">
				<div class="card-title col s7">Addons</div>	
				</div>
				<ul class="posts row">`

		_, err := html.WriteString(open)
		if err != nil {
			log.Println("Error writing open html to addon html view:", err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				log.Println(err)
				return
			}

			res.Write(errView)
			return
		}

		_, err = html.Write(list.Bytes())
		if err != nil {
			log.Println("Error writing list bytes to addon html view:", err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				log.Println(err)
				return
			}

			res.Write(errView)
			return
		}

		_, err = html.WriteString(`</ul></div></div>`)
		if err != nil {
			log.Println("Error writing close html to addon html view:", err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				log.Println(err)
				return
			}

			res.Write(errView)
			return
		}

		if html.Len() == 0 {
			_, err := html.WriteString(`<p>No addons available.</p>`)
			if err != nil {
				log.Println("Error writing default addon html to admin view:", err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					log.Println(err)
					return
				}

				res.Write(errView)
				return
			}
		}

		view, err := Admin(html.Bytes())
		if err != nil {
			log.Println("Error writing addon html to admin view:", err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				log.Println(err)
				return
			}

			res.Write(errView)
			return
		}

		res.Write(view)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		id := req.PostFormValue("id")
		action := strings.ToLower(req.PostFormValue("action"))

		at, ok := addon.Types[id]
		if !ok {
			log.Println("Error: no addon type found for:", id)
			log.Println(err)
			res.WriteHeader(http.StatusNotFound)
			errView, err := Error404()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		b, err := db.Addon(id)
		if err == db.ErrNoAddonExists {
			log.Println(err)
			res.WriteHeader(http.StatusNotFound)
			errView, err := Error404()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		adn := at()
		err = json.Unmarshal(b, adn)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		h, ok := adn.(item.Hookable)
		if !ok {
			log.Println("Addon", adn, "does not implement the item.Hookable interface or embed item.Item")
			return
		}

		switch action {
		case "enable":
			err := h.BeforeEnable(res, req)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			err = addon.Enable(id)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			err = h.AfterEnable(res, req)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

		case "disable":
			err := h.BeforeDisable(res, req)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			err = addon.Disable(id)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			err = h.AfterDisable(res, req)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}
		default:
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		http.Redirect(res, req, req.URL.String(), http.StatusFound)

	default:
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			log.Println(err)
			return
		}

		res.Write(errView)
		return
	}
}

func addonRestHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		id := req.FormValue("id")

		data, err := db.Addon(id)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		_, ok := addon.Types[id]
		if !ok {
			log.Println("Addon: ", id, "is not found in addon.Types map")
			res.WriteHeader(http.StatusNotFound)
			errView, err := Error404()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		m, err := addon.Manage(data, id)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		addonView, err := Admin(m)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(addonView)

	case http.MethodPost:
		// save req.Form
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		name := req.FormValue("addon_name")
		id := req.FormValue("addon_reverse_dns")

		at, ok := addon.Types[id]
		if !ok {
			log.Println("Error: addon", name, "has no record in addon.Types map at", id)
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// if Hookable, call BeforeSave prior to saving
		h, ok := at().(item.Hookable)
		if ok {
			err := h.BeforeSave(res, req)
			if err != nil {
				log.Println("Error running BeforeSave method in addonHandler for:", id, err)
				return
			}
		}

		err = db.SetAddon(req.Form, at())
		if err != nil {
			log.Println("Error saving addon:", name, err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		http.Redirect(res, req, "/admin/addon?id="+id, http.StatusFound)

	default:
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error405()
		if err != nil {
			log.Println(err)
			return
		}

		res.Write(errView)
		return
	}
}
