package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/item"
	"github.com/nfnt/resize"
)

// ErrNoAuth should be used to report failed auth requests
var RestErrNoAuth = errors.New("Auth failed for request")

// deprecating from API, but going to provide code here in case someone wants it
func __typesHandler(res http.ResponseWriter, req *http.Request) {
	var types = []string{}
	for t, fn := range item.Types {
		if !hide(res, req, fn()) {
			types = append(types, t)
		}
	}

	j, err := toJSON(types)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, req, j)
}

func contents(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	it, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if hide(res, req, it()) {
		return
	}

	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if q.Get("count") == "" {
			count = 10
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	order := strings.ToLower(q.Get("order")) // string: sort order of posts by timestamp ASC / DESC (DESC default)
	if order != "asc" {
		order = "desc"
	}

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}
	// assert hookable
	get := it()
	hook, ok := get.(item.Hookable)
	if !ok {
		logger.Warn("[Response] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	selfContent := hook.EnableOwnerCheck()
	var total int
	var bb [][]byte
	if selfContent {
		userEmail, err := getEmailFromCookie(req)
		if err != nil {
			logger.Warn("Error when extract use cookie")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		total, bb = db.QueryByFieldValue(t+"__sorted", "email", userEmail, opts)
	} else {
		total, bb = db.Query(t+"__sorted", opts)
	}

	var result = []json.RawMessage{}
	for i := range bb {
		result = append(result, bb[i])
	}

	retData := make([]map[string]interface{}, 0)
	for i := range result {
		m := map[string]interface{}{}
		err := json.Unmarshal(result[i], &m)
		if err == nil {
			retData = append(retData, m)
		} else {
			logger.Warn("Convert error.")
		}
	}
	fields, hasSubContent := hook.EnableSubContent()

	if hasSubContent {

		logger.Debug("Now process sub-content ")
		for kk := range retData {
			for index := range fields {
				fieldname := fields[index]
				data, err := db.GetSubContent(t+":"+fmt.Sprint(retData[kk]["id"]), fieldname)
				fmt.Println(t + ":" + fmt.Sprint(retData[kk]["id"]))
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

	j, err := fmtMAP(retData...)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, it(), j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// hook before response
	j, err = hook.BeforeAPIResponse(res, req, j)

	if err != nil {
		logger.Error("[Response] error calling BeforeAPIResponse:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// fmt.Println(total)
	r := bytes.NewReader(j)
	resp := make(map[string][]json.RawMessage, 0)
	ddd := json.NewDecoder(r)
	//fmt.Println(ddd)
	err = ddd.Decode(&resp)

	dada := resp["data"]

	p := 0

	if count > 0 {
		p = total / count
	} else {
		p = 1
	}
	meta := MetaData{
		Total:     uint(total),
		PageCount: p,
		Page:      offset,
		Order:     order,
		PageSize:  count, //-1 means all
	}

	ReturnStructDataBytes(res, req, dada, meta)

	//sendData(res, req, j)

	// hook after response
	err = hook.AfterAPIResponse(res, req, j)
	if err != nil {
		logger.Warn("[Response] error calling AfterAPIResponse:", err)
		return
	}

}

func content(res http.ResponseWriter, req *http.Request) {
	logger.Debug("Get content by id, From ", GetIP(req))
	q := req.URL.Query()
	id := q.Get("id")
	t := q.Get("type")
	slug := q.Get("slug")

	if slug != "" {
		contentHandlerBySlug(res, req)
		return
	}

	if t == "" || id == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	pt, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	post, err := db.Content(t + ":" + id)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	p := pt()
	err = json.Unmarshal(post, p)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if hide(res, req, p) {
		return
	}

	retdata := map[string]interface{}{}
	//err = json.Unmarshal(data, post)
	err = json.Unmarshal(post, &retdata)
	if err != nil {
		logger.Error(err.Error())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	// assert hookable
	get := p
	hook, ok := get.(item.Hookable)
	if !ok {
		logger.Debug("[Response] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if hook.EnableOwnerCheck() { //check the content is user belong
		ee := q.Get("email") //check url to see if have email parameter
		if ee == "" {
			email, ok := getEmailFromCookie(req) //otherwise check the cookie
			if ok == nil {
				theemail := retdata["email"].(string)
				if email != theemail {
					logger.Warn("Try to get other user content", email, theemail)
					res.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
			} else {
				logger.Warn("User information in cookie have problem ")
				res.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			theemail := retdata["email"].(string)
			if ee != theemail {
				logger.Warn("Try to get other user content", ee, theemail)
				res.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		}
	}
	if ok {
		// hook before response
		fields, hasSubContent := hook.EnableSubContent()
		if hasSubContent {

			logger.Debug("Now process sub-content ")

			for index := range fields {
				fieldname := fields[index]
				data, err := db.GetSubContent(t+":"+id, fieldname)
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

	//push(res, req, p, post)

	//j, err := fmtJSON(json.RawMessage(post))
	j, err := fmtMAP(retdata)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, p, j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// hook before response
	j, err = hook.BeforeAPIResponse(res, req, j)
	if err != nil {
		logger.Error("[Response] error calling BeforeAPIResponse:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	r := bytes.NewReader(j)
	resp := make(map[string][]json.RawMessage, 0)
	ddd := json.NewDecoder(r)
	//fmt.Println(ddd)
	err = ddd.Decode(&resp)

	dada := resp["data"]

	//ret := []json.RawMessage{}
	//ret = append(ret, dada)
	///	sendData(res, req, j)

	RenderJSON(res, req, ReturnDataBytes{
		RetCode: 0,
		Msg:     "ok",
		Data:    dada,
	})
	// hook after response
	err = hook.AfterAPIResponse(res, req, j)
	if err != nil {
		logger.Error("[Response] error calling AfterAPIResponse:", err)
		return
	}
}

func contentBySlug(res http.ResponseWriter, req *http.Request) {
	slug := req.URL.Query().Get("slug")

	if slug == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// lookup type:id by slug key in __contentIndex
	t, post, err := db.ContentBySlug(slug)
	if err != nil {
		logger.Error("Error finding content by slug:", slug, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	it, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	p := it()
	err = json.Unmarshal(post, p)
	if err != nil {
		logger.Error(err)
		return
	}

	if hide(res, req, p) {
		return
	}

	retdata := map[string]interface{}{} // 返回数据结构
	//err = json.Unmarshal(data, post)
	err = json.Unmarshal(post, &retdata)
	if err != nil {
		logger.Error(err.Error())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	id := fmt.Sprint(retdata["id"])
	// assert hookable
	get := p
	hook, ok := get.(item.Hookable)
	if !ok {
		logger.Debug("[Response] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if ok {
		// hook before response
		fields, hasSubContent := hook.EnableSubContent()
		if hasSubContent {

			logger.Debug("Now process sub-content ")

			for index := range fields {
				fieldname := fields[index]

				data, err := db.GetSubContent(t+":"+id, fieldname)
				//fmt.Println(t + ":" + id)
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
	//push(res, req, p, post)

	j, err := fmtMAP(retdata)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, p, j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// hook before response
	j, err = hook.BeforeAPIResponse(res, req, j)
	if err != nil {
		logger.Error("[Response] error calling BeforeAPIResponse:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	r := bytes.NewReader(j)
	resp := make(map[string][]json.RawMessage, 0)
	ddd := json.NewDecoder(r)
	//fmt.Println(ddd)
	err = ddd.Decode(&resp)

	dada := resp["data"]

	RenderJSON(res, req, ReturnDataBytes{
		RetCode: 0,
		Msg:     "ok",
		Data:    dada,
	})
	//sendData(res, req, j)

	// hook after response
	err = hook.AfterAPIResponse(res, req, j)
	if err != nil {
		logger.Error("[Response] error calling AfterAPIResponse:", err)
		return
	}

}

func uploads(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	slug := req.URL.Query().Get("slug")
	if slug == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	upload, err := db.UploadBySlug(slug)
	if err != nil {
		logger.Error("Error finding upload by slug:", slug, err)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	it := func() interface{} {
		return new(item.FileUpload)
	}

	push(res, req, it(), upload)

	j, err := fmtJSON(json.RawMessage(upload))
	if err != nil {
		logger.Error("Error fmtJSON on upload:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, it(), j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, req, j)
}

// return the pic content
func getMedia(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	id := q.Get("id") // int: multiplier of count for pagination (0 default)
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	t := db.DB__uploads // upload db

	contentbyte, err := db.Upload(t + ":" + id)
	if err != nil {
		logger.Errorf("The Media is is not found %s,error:%s", id, err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	item := make(map[string]interface{})

	err = json.Unmarshal(contentbyte, &item)
	if err != nil {
		logger.Errorf("Error unmarshal json into:%s", err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
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
	//logger.Debugf("The file  %s being read \n", mediaFilename)
	//logger.Debugf(strings.TrimPrefix(fff, "/api/uploads"))
	dat, err := ioutil.ReadFile(mediaFilename)
	if err != nil {
		logger.Error("Couldn't read file content.")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//content_type := mime.TypeByExtension(file_ext)

	//logger.Info("content type is %s", ctype)

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
			logger.Error("input parameter w is error:", err.Error())
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
			logger.Error("input parameter h is error :", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		is_height_set = true
	}

	//logger.Debugf("width and height is [%d, %d], status [%t, %t]", width, height, is_width_set, is_height_set)

	if is_width_set || is_height_set {
		var (
			original_image image.Image
			new_image      image.Image
		)
		if original_image, _, err = image.Decode(bytes.NewReader(dat)); nil != err {
			logger.Errorf("image decode error! %v", err)
			goto LABEL_IMAGE_HANDLE_FINISHED
		}

		new_image = resize.Resize(uint(width), uint(height), original_image, resize.Lanczos3)
		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, new_image, nil); nil != err {
			logger.Errorf("image encode error! %v", err)
			goto LABEL_IMAGE_HANDLE_FINISHED
		}
		dat = buf.Bytes()

		r.Header.Set("Content-Type", "image/jpeg")
	}

LABEL_IMAGE_HANDLE_FINISHED:

	w.Write(dat)
	logger.Debugf("Media(id %s) sent!", id)
	w.WriteHeader(http.StatusOK)
	//http.Redirect(w, r, "/api/uploads"+fmt.Sprint(item["path"]), http.StatusFound)
}
