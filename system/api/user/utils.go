package user

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
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

func stripPrefix(prefix string, h http.Handler) http.Handler {
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
