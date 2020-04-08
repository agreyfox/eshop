package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"reflect"
	"strings"
)

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

func fmtJSON(data ...json.RawMessage) ([]byte, error) {
	var msg = []json.RawMessage{}
	for _, d := range data {
		msg = append(msg, d)
	}

	resp := map[string][]json.RawMessage{
		"data": msg,
	}

	var buf = &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(resp)
	if err != nil {
		log.Println("Failed to encode data to JSON:", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

//将interface 简单传回
func renderJSON(w http.ResponseWriter, r *http.Request, data interface{}) (int, error) {

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

// 按照标准返回格式返回数据
func returnStructData(w http.ResponseWriter, r *http.Request, data []map[string]interface{}, meta MetaData) (int, error) {
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

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

// print pretty map[string]internface output
func PrettyPrint(obj map[string]interface{}) {
	prettyJSON, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Fatal("Failed to generate json", err)
	}
	fmt.Println("===================================================================")
	fmt.Printf("\t\t%s\n", string(prettyJSON))
	fmt.Println("===================================================================")
}

func PrettyPrintJson(obj interface{}) {
	res2B, _ := json.Marshal(obj)
	fmt.Println(string(res2B))
}

func getJsonFromBody(req *http.Request) map[string]interface{} {
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
		logger.Debug("jons data looks like bad")
		logger.Debugf("%+v", err)
		return nil
	}
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

func sendEmail(server, from, to, password, subject, body string) error {
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
