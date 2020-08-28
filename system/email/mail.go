package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/logs"
	"go.uber.org/zap"
)

var (
	MailServer                        = "mail.smtp2go.com"
	MailServerHTTP                    = "https://api.smtp2go.com/v3"
	MailKey                           = "api-E408790EAED711EA8BC0F23C91C88F4E"
	MailUser                          = "support@bk.cloudns.cc"
	logger         *zap.SugaredLogger = logs.Log.Sugar()
)

const api_root_env string = "ESHOP_MAIL_ROOT"
const api_key_env string = "ESHOP_MAIL_KEY"
const api_header string = "X-Smtp2go-Api"
const api_version_header string = "X-Smtp2go-Api-Version"
const api_key_header string = "X-Smtp2go-Api-Key"

var api_key_regex *regexp.Regexp = regexp.MustCompile("^api-[a-zA-Z0-9]{32}$")

type (
	Email struct {
		ApiKey   string   `json:"api_key`
		From     string   `json:"sender"`
		To       []string `json:"to"`
		Subject  string   `json:"subject"`
		TextBody string   `json:"text_body"`
		HtmlBody string   `json:"html_body"`
	}

	SendAsyncResult struct {
		Error  error
		Result *Smtp2goApiResult
	}

	Smtp2goApiResult struct {
		RequestId string                `json:"request_id"`
		Data      Smtp2goApiResult_Data `json:"data"`
	}

	Smtp2goApiResult_Data struct {
		Succeeded int      `json:"succeeded,omitempty"`
		Failed    int      `json:"failed,omitempty"`
		Failures  []string `json:"failures,omitempty"`
		EmailID   string   `json:"email_id,omitempty"`
	}

	IncorrectAPIKeyFormatError struct{ found string }
	MissingRequiredFieldError  struct{ field string }
	RequestError               struct{ err error }
	EndpointError              struct{ err error }
	InvalidJSONError           struct{ err error }
)
type MissingAPIKeyError string

func (f MissingAPIKeyError) Error() string {
	return fmt.Sprintf("The %s environment variable was not found, please export it or set it in code prior to api calls", api_key_env)
}

func (f IncorrectAPIKeyFormatError) Error() string {
	return fmt.Sprintf("The value of SMTP2GO_API_KEY %s does not match the api key format of ^api-[a-zA-Z0-9]{{32}}$, please correct it", f.found)
}

func (f MissingRequiredFieldError) Error() string {
	return fmt.Sprintf("%s is a required field.", f.field)
}

func (f RequestError) Error() string {
	return fmt.Sprintf("Something went wrong with the request: %s.", f.err)
}

func (f EndpointError) Error() string {
	return fmt.Sprintf("Something went wrong with the request: %s.", f.err)
}

func (f InvalidJSONError) Error() string {
	return fmt.Sprintf("Unable to serialise request into valid JSON: %s", f.err)
}

func api_request(endpoint string, request io.Reader) (*Smtp2goApiResult, error) {

	// grab the api_root_env, set it if it's empty
	api_root, found := os.LookupEnv(api_root_env)
	if !found || len(api_root) == 0 {
		r, err := db.Config("email_domain")
		if err == nil {
			api_root = string(r)
			MailServerHTTP = string(r)
		}
		api_root = MailServerHTTP

	}

	// grab the api_key env
	api_key, found := os.LookupEnv(api_key_env)
	if !found || len(api_key) == 0 {
		r, err := db.Config("admin_email")
		if err == nil {
			MailUser = string(r)
		}
		r, err = db.Config("email_password")
		if err == nil {
			MailKey = string(r)
		}
		api_key = MailKey
	}

	// check if the api key is valid
	if !api_key_regex.MatchString(api_key) {
		return nil, &IncorrectAPIKeyFormatError{found: api_key}
	}
	logger.Infof("Email api called with url:", api_root+"/"+endpoint)

	// create the http request client
	client := &http.Client{}
	req, err := http.NewRequest("POST", api_root+"/"+endpoint, request)
	if err != nil {
		return nil, &RequestError{err: err}
	}

	// add the headers
	req.Header.Add(api_header, "smtp2go-go")
	req.Header.Add(api_version_header, "0.1")
	req.Header.Add(api_key_header, api_key)
	req.Header.Add("Content-Type", "applications/json")

	// make the request and grab the response
	res, err := client.Do(req)
	if err != nil {
		return nil, &RequestError{err: err}
	}

	// otherwise unmarshal the data into a result object

	ret2 := new(Smtp2goApiResult)
	//ret := map[string]interface{}{}
	err = json.NewDecoder(res.Body).Decode(ret2)
	if err != nil {
		return nil, &InvalidJSONError{err: err}
	}
	logger.Debugf("%v\n", ret2)
	// finally return the result object
	return ret2, nil
}

func Send(e *Email) (*Smtp2goApiResult, error) {

	// check that we have From data
	/* 	if len(e.From) == 0 {
		return nil, MissingRequiredFieldError{field: "From"}
	} */

	// check that we have To data
	if len(e.To) == 0 {
		return nil, MissingRequiredFieldError{field: "To"}
	}

	// check that we have Subject data
	if len(e.Subject) == 0 {
		return nil, MissingRequiredFieldError{field: "Subject"}
	}

	// check that we have TextBody data
	if len(e.TextBody) == 0 {
		return nil, MissingRequiredFieldError{field: "TextBody"}
	}
	e.ApiKey = MailKey
	e.From = MailUser
	fmt.Printf("%v", e)
	// if we get here we have enough information to send
	request_json, err := json.Marshal(e)
	if err != nil {
		return nil, &InvalidJSONError{err: err}
	}

	// now call to api_request in core to do the http request
	res, err := api_request("/email/send", bytes.NewReader(request_json))
	if err != nil {
		return res, err
	}

	return res, nil
}

func SendAsync(e *Email) chan *SendAsyncResult {

	c := make(chan *SendAsyncResult)
	go func() {
		res, err := Send(e)
		if err != nil {
			c <- &SendAsyncResult{Error: err}
		}
		c <- &SendAsyncResult{Result: res}
	}()
	return c
}

/*
{
    "api_key": "api-40246460336B11E6AA53F23C91285F72",
    "to": ["Test Person <test@example.com>"],
    "sender": "Test Persons Friend <test2@example.com>",
    "subject": "Hello Test Person",
    "text_body": "You're my favorite test person ever",
    "html_body": "<h1>You're my favorite test person ever</h1>",
    "custom_headers": [
      {
        "header": "Reply-To",
        "value": "Actual Person <test3@example.com>"
      }
    ],
    "attachments": [
        {
            "filename": "test.pdf",
            "fileblob": "--base64-data--",
            "mimetype": "application/pdf"
        },
        {
            "filename": "test.txt",
            "fileblob": "--base64-data--",
            "mimetype": "text/plain"
        }
    ]
}
return data
{
  "request_id": "aa253464-0bd0-467a-b24b-6159dcd7be60",
  "data":
  {
    "succeeded": 1,
    "failed": 0,
    "failures": [],
    "email_id": "1er8bV-6Tw0Mi-7h"
  }
}
*/

// ues smtp2go api to send email
/* func sendMail(from, to, subject, body string) bool {
	rreq := SMTP2GOReq{
		APIKEY:       MailKey,
		To:           []string{},
		Sender:       from,
		Subject:      subject,
		HTMLBody:     body,
		CustomHeader: []CustomHeader{},
		Attachments:  []Attachment{},
	}

	rreq.CustomHeader = append(rreq.CustomHeader, CustomHeader{
		Header: "Reply-To",
		Value:  MailUser,
	})
	rreq.To = append(rreq.To, to)
	return true
}
*/
