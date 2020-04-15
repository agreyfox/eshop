package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"net/url"
	"reflect"
	"strings"

	"github.com/agreyfox/eshop/system/admin/user"
	"github.com/nilslice/jwt"
)

type MetaOfRecorder struct {
	Total     uint32 `json:"total,omitempty"`
	PageCount uint16 `json:"pageCount,omitempty"`
	Page      uint16 `json:"page,omitempty"`
}

type RetUser struct {
	RetCode int8           `json:"retCode"`
	Data    interface{}    `json:"data,omitempty"`
	Meta    MetaOfRecorder `json:"meta,omitempty" `
	Msg     string         `json:msg`
}

type MetaData struct {
	Total     uint   `json:"total,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
	PageCount int    `json:"pageCount,omitempty"`
	Page      int    `json:"page,omitempty"`
	Order     string `json:"order,omitempty"`
}

type ReturnData struct {
	RetCode int           `json:"retCode"`
	Msg     string        `json:"message"`
	Data    []interface{} `json:"data,omitempty"`
	Meta    MetaData      `json:"meta,omitempty"`
}

type ReturnDataBytes struct {
	RetCode int               `json:"retCode"`
	Msg     string            `json:"message"`
	Data    []json.RawMessage `json:"data,omitempty"`
	Meta    MetaData          `json:"meta,omitempty"`
}

var (
	PENDINGSuffix    = "__pending"
	SORTSuffix       = "__sorted"
	LQCMStoken       = "lqcms_token"
	LQCMSRequestJson = "lqcms_json" // not use right now
	// ErrNoAuth should be used to report failed auth requests
	ErrNoAuth   = errors.New("Auth failed for request")
	ErrNoCookie = errors.New("Request without cookie")
	ErrBadToken = errors.New("Token Wrong")
)

func RenderJSON(w http.ResponseWriter, r *http.Request, data interface{}) (int, error) {
	marsh, err := json.Marshal(data)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

// 将数据按照 []byte 返回
func ReturnStructDataBytes(w http.ResponseWriter, r *http.Request, data []json.RawMessage, meta MetaData) (int, error) {

	marsh, err := json.Marshal(ReturnDataBytes{
		RetCode: 0,
		Msg:     "ok",
		Data:    data,
		Meta:    meta,
	})

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

// 按照标准返回格式返回数据
func ReturnStructData(w http.ResponseWriter, r *http.Request, data []map[string]interface{}, meta MetaData) (int, error) {
	s := reflect.ValueOf(data)
	value := []interface{}{}
	if s.Kind() != reflect.Slice {
		logger.Debug("data is not array ")
		value = append(value, data)
	} else {
		for i := range data {
			value = append(value, data[i])
		}
	}

	//ret := make([]interface{}, s.Len())

	marsh, err := json.Marshal(ReturnData{
		RetCode: 0,
		Msg:     "ok",
		Data:    value,
		Meta:    meta,
	})

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

// get json from request body
func GetJsonFromBody(req *http.Request) map[string]interface{} {

	bodyBytes, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		logger.Debugf("failed read Read response body: %v", err)
		//logger.Debug(body)
		return map[string]interface{}{}
	}
	var t map[string]interface{}

	//fmt.Println(body)

	if err = json.Unmarshal(bodyBytes, &t); err != nil {

		return map[string]interface{}{}
	}
	req.Header.Add("lqcms_json", string(bodyBytes[:]))
	//logger.Debug("add object to header")
	return t

}

//convert map[string]interface to url.Values to save to db
// url.Values is map[string][]string
func formatData(data map[string]interface{}) url.Values {
	retdata := url.Values{}
	for key, value := range data {

		onedata, o := value.([]string)
		if !o {
			retdata[key] = []string{fmt.Sprint(value)}
		} else {
			retdata[key] = onedata
		}
	}

	return retdata
}

func StripPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = p
		h.ServeHTTP(w, r2)
	})
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
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

func SendEmail(server, from, to, password, subject, body string) error {
	hp := strings.Split(server, ":")
	sub := subject
	content := body
	mailList := strings.Split(to, ",")

	auth := smtp.PlainAuth(
		"",
		from,
		password,
		hp[0],
		//"smtp.gmail.com",
	)
	logger.Debug(auth)
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		server,
		auth,
		from,
		mailList,
		[]byte(sub+content),
	)
	return err
}

func getEmailFromCookie(req *http.Request) (string, error) {
	userEmail := ""
	cookie, err := req.Cookie(LQCMStoken)
	if err != nil {

		return "", ErrNoCookie
	}
	// validate it and allow or redirect request
	token := cookie.Value
	if jwt.Passes(token) {
		clienInfo := jwt.GetClaims(token)
		userEmail = clienInfo["user"].(string)
	} else {
		return "", ErrBadToken
	}
	return userEmail, nil
}

/*
func DecodeJwt(token string) map[string]interface{} {
	data := strings.Split(token, ".")
	fmt.Printf("%v", data)
	var target map[string]interface{}
	if len(data) > 1 {
		dt, _ := base64.StdEncoding.DecodeString(data[1])
		fmt.Println(dt)
		err := json.Unmarshal(dt, &target)
		fmt.Println(err)
	}
	fmt.Println(target)
	return target
} */
