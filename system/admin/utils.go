package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/item"
)

type (
	/* 	SMTP2GOReq struct {
	   		APIKEY       string         `json:"api_key"`
	   		To           []string       `json:"to"`
	   		Sender       string         `json:"sender"`
	   		Subject      string         `json:"subject,omitempty"`
	   		TextBody     string         `json:"text_body,omitempty"`
	   		HTMLBody     string         `json:"html_body,omitempty"`
	   		CustomHeader []CustomHeader `json:"custom_header,omitempty"`
	   		Attachments  []Attachment   `json:"attachment,omitempty"`
	   	}

	   	CustomHeader struct {
	   		Header string `json:"header,omitempty"`
	   		Value  string `json:"value,omitempty"`
	   	}

	   	Attachment struct {
	   		FileName string `json:"filename,omitempty"`
	   		FileBlob string `json:"fileblob,omitempty"`
	   		MimeType string `json:"mimetype,omitempty"`
	   	} */
	MetaData struct {
		Total     uint   `json:"total,omitempty"`
		PageSize  int    `json:"pageSize,omitempty"`
		PageCount int    `json:"pageCount,omitempty"`
		Page      int    `json:"page,omitempty"`
		Order     string `json:"order,omitempty"`
	}

	ReturnData struct {
		RetCode int           `json:"retCode"`
		Msg     string        `json:"message"`
		Data    []interface{} `json:"data,omitempty"`
		Meta    MetaData      `json:"meta,omitempty"`
	}
)

var (
	PENDINGSuffix = "__pending"
	SORTSuffix    = "__sorted"
)

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
		logger.Error("Failed to encode data to JSON:", err)
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
		logger.Fatal("Failed to generate json", err)
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
} */

//return system content and structure with order
func getContentsStruct() (ret []byte) {
	retdata := map[string]interface{}{}
	for key, it := range item.Types {
		//fmt.Println(key)
		obj := it()
		s, ok := obj.(item.ContentStructable)
		if ok {
			retdata[key] = s.ContentStruct()
		}
	}
	data, err := json.Marshal(retdata)
	if err != nil {
		logger.Error(err)
		return []byte{}
	}
	//fmt.Println(string(data[:]))
	return data
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
type Mail struct {
	user   string
	passwd string
}

//初始化用户名和密码
func NewMailClient(u string, p string) Mail {
	temp := Mail{user: u, passwd: p}
	return temp
}

func check(err error) {
	if err != nil {
		logger.Error(err)
	}
}

//标题 文本 目标邮箱
func (m Mail) Send(title string, text string, toId string) {
	auth := smtp.PlainAuth("", m.user, m.passwd, MailServer)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         MailServer,
	}

	conn, err := tls.Dial("tcp", MailServer+":587", tlsconfig)

	client, err := smtp.NewClient(conn, MailServer)
	check(err)

	if err = client.Auth(auth); err != nil {
		logger.Error(err)
		return
	}

	if err = client.Mail(m.user); err != nil {
		logger.Error(err)
		return
	}

	if err = client.Rcpt(toId); err != nil {
		logger.Error(err)
		return
	}

	w, err := client.Data()
	check(err)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.user, toId, title, text)

	_, err = w.Write([]byte(msg))
	check(err)

	err = w.Close()
	check(err)

	client.Quit()
}
*/
