package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/agreyfox/eshop/system/admin/user"
	"github.com/agreyfox/eshop/system/db"
	"github.com/nilslice/jwt"
)

type MetaOfRecorder struct {
	Total     uint32 `json:"total,omitempty"`
	PageCount uint16 `json:"pageCount,omitempty"`
	Page      uint16 `json:"page,omitempty"`
}

type RetUser struct {
	RetCode        int8           `json:"retCode"`
	Data           interface{}    `json:"data,omitempty"`
	Meta           MetaOfRecorder `json:"meta,omitempty" `
	Msg            string         `json:msg`
	DefaultCountry string         `json:default_country,omitempty`
	Country        []string       `json:"country,omitempty"`
	Currency       []string       `json:"currency,omitempty"`
	Buttons        []string       `json:"buttons,omitempty"`
	SocialType     string         `json:"social_type,omitempty"`
	SocialLink     string         `json:"social_link,omitempty"`
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

type UpdateUserRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password,omitempty"`
	Social      string `json:"social_link,omitempty"`
	Type        string `json:"social_type,omitempty"`
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
		logger.Error("Error:", err)
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

func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// RemoteIP 通过 RemoteAddr 获取 IP 地址， 只是一个快速解析方法。
func RemoteIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// IPString2Long 把ip字符串转为数值
func IPString2Long(ip string) (uint, error) {
	b := net.ParseIP(ip).To4()
	if b == nil {
		return 0, errors.New("invalid ipv4 format")
	}

	return uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24, nil
}

// Long2IPString 把数值转为ip字符串
func Long2IPString(i uint) (string, error) {
	if i > math.MaxUint32 {
		return "", errors.New("beyond the scope of ipv4")
	}

	ip := make(net.IP, net.IPv4len)
	ip[0] = byte(i >> 24)
	ip[1] = byte(i >> 16)
	ip[2] = byte(i >> 8)
	ip[3] = byte(i)

	return ip.String(), nil
}

// IP2Long 把net.IP转为数值
func IP2Long(ip net.IP) (uint, error) {
	b := ip.To4()
	if b == nil {
		return 0, errors.New("invalid ipv4 format")
	}

	return uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24, nil
}

// Long2IP 把数值转为net.IP
func Long2IP(i uint) (net.IP, error) {
	if i > math.MaxUint32 {
		return nil, errors.New("beyond the scope of ipv4")
	}

	ip := make(net.IP, net.IPv4len)
	ip[0] = byte(i >> 24)
	ip[1] = byte(i >> 16)
	ip[2] = byte(i >> 8)
	ip[3] = byte(i)

	return ip, nil
}
func HasLocalIPddr(ip string) bool {
	return HasLocalIP(net.ParseIP(ip))
}

func HasLocalIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	return ip4[0] == 10 || // 10.0.0.0/8
		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
}

func ClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" && !HasLocalIPddr(ip) {
			return ip
		}
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" && !HasLocalIPddr(ip) {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		if !HasLocalIPddr(ip) {
			return ip
		}
	}

	return ""
}

// Use sofescate way to get IP address.
func GetIP(r *http.Request) string {
	/* forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr */
	ip := ClientPublicIP(r)
	if ip == "" {
		ip = ClientIP(r)
	}
	return ip
}

/*
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
*/
// Auth is HTTP middleware to ensure the request has proper token credentials
func CustomerAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		if user.IsValid(req) {
			next.ServeHTTP(res, req)
		} else {
			res.WriteHeader(http.StatusForbidden)
			logger.Errorf("Action %s without user permission:", req.RequestURI)
			return
		}
	})
}

// get all currency list in systrem
func getContentList(name string) []string {
	ret := []string{}
	contentBuff := db.ContentAll(name)
	for i := range contentBuff {
		ret = append(ret, string(contentBuff[i]))
	}
	return ret
}

/*
func SendEmail(server, from, to, password, subject, body string) error {
	hp := strings.Split(server, ":")
	sub := subject
	content := body
	mailList := strings.Split(to, ",")

	auth := smtp.PlainAuth(
		"",
		from,
		password,
		hp[0],408006570@qq.com
		server,
		auth,
		from,
		mailList,
		[]byte(sub+content),
	)
	return err
} */

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

// to create new password, 8 letter
func GeneratePassword() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789-+#$%")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // E.g. "ExcbsVQs"
	return str
}
