package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// generator short orderid  based on customer request
func GetShortOrderID() string {
	now := time.Now()
	year, month, _ := time.Now().Date()

	mil := int64(time.Nanosecond) * now.UnixNano() / int64(time.Millisecond)
	str := fmt.Sprint(mil)
	ret := fmt.Sprint(year) + fmt.Sprint(int(month)) + str[len(str)-6:len(str)-1]
	return ret

}

// print pretty map[string]internface output
func PrettyPrint(obj interface{}) {

	prettyJSON, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		logger.Fatal("Failed to generate json", err)
	}
	fmt.Println("===================================================================")
	fmt.Printf("%s\n", string(prettyJSON))
	fmt.Println("===================================================================")
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

func GetJSONFromBody(req *http.Request) map[string]interface{} {
	var body string
	bodyBytes, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		body = fmt.Sprintf("failed read Read response body: %v", err)
		logger.Debug(body)
		return nil
	}
	var t map[string]interface{}

	//fmt.Println(body)

	if err = json.Unmarshal(bodyBytes, &t); err != nil {
		logger.Debug("json data looks like bad")
		logger.Debugf("%+v", err)
		return nil
	}
	return t
}

//将interface 简单传回
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

func GetBinaryDataFromBody(req *http.Request) []byte {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(req.Body)
	}
	// Restore the io.ReadCloser to its original state
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
