package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/item"
	"github.com/agreyfox/eshop/system/search"
)

func searchContentHandler(res http.ResponseWriter, req *http.Request) {

	qs := req.URL.Query()

	t := qs.Get("type")
	// type must be set, future version may compile multi-type result set
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	it, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if hide(res, req, it()) {
		return
	}

	q, err := url.QueryUnescape(qs.Get("q"))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// q must be set
	if q == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	count, err := strconv.Atoi(qs.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if qs.Get("count") == "" {
			count = 10
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	offset, err := strconv.Atoi(qs.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if qs.Get("offset") == "" {
			offset = 0
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// execute search for query provided, if no index for type send 404
	matches, err := search.TypeQuery(t, q, count, offset)
	if err == search.ErrNoIndex {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("[search] Error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// respond with json formatted results
	bb, err := db.ContentMulti(matches)
	if err != nil {
		log.Println("[search] Error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// if we have matches, push the first as its matched by relevance
	if len(bb) > 0 {
		push(res, req, it(), bb[0])
	}

	var result = []json.RawMessage{}
	for i := range bb {
		result = append(result, bb[i])
	}

	j, err := fmtJSON(result...)
	if err != nil {
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

// to describer the search critireal.
type ResultFilter struct {
	KeyName string `json:"keyName"`
	Value   string `json:"value"`
	Include bool   `json:"include"`
}

// for search advance content
func advSearchContent(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	t := q.Get("type")
	search := q.Get("q")
	logger.Infof("Search content %s with %s from %s", t, search, GetIP(r))
	status := q.Get("status")
	regexsearch := q.Get("r")
	starttime := q.Get("start") // this is base on time query
	endtime := q.Get("end")     // this is base on time query
	filter := q.Get("filter")   // add 2020/10/27 增加filter，在查询结果上加上filter，必须冒号分割,key:value,key !开头就是否（包含的不要）
	var filterObj *ResultFilter
	if len(filter) >= 0 {
		filterStr := strings.Split(filter, ":")
		if len(filterStr) != 2 {
			filterStr = []string{}
			filterObj = nil
		} else {
			filterObj = &ResultFilter{}
			if filterStr[0][0] == '!' {
				filterObj.KeyName = filterStr[0][1:]
				filterObj.Value = filterStr[1]
				filterObj.Include = false
			} else {
				filterObj.KeyName = filterStr[0]
				filterObj.Value = filterStr[1]
				filterObj.Include = true
			}
		}
	} else {
		filterObj = nil
	}
	var checkTime bool
	checkTime = false
	var stime, etime uint64
	var err error
	if len(starttime) > 0 || len(endtime) > 0 {
		checkTime = true
		stime, err = strconv.ParseUint(starttime, 10, 64)
		if err != nil {
			stime = 0
			logger.Warn("Search no start time")
		} else {
			logger.Debug("Search start time is ", stime)
		}
		if err != nil {
			etime = 0
			logger.Warn("Search no end time")
		} else {
			logger.Debug("Search end time is ", etime)
		}
	} else {
		logger.Debug("No Time range search present ")
		stime = 0
		etime = 0
	}

	var specifier string

	if t == "" || (search == "" && regexsearch == "" && !checkTime) {
		logger.Debugf("Search parameter missing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if status == "pending" {
		specifier = "__" + status
	}
	var posts [][]byte
	if !checkTime {
		logger.Debug("Get all content:")
		posts = db.ContentAll(t + specifier)
	} else {
		logger.Debug("Get content based on time frame:", stime, etime)
		posts = db.ContentByUpdatedTime(t+specifier, stime, etime)
	}

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
			if filterObj != nil {
				value := fmt.Sprint(item[filterObj.KeyName])
				if filterObj.Include {
					if value == filterObj.Value {
						retData = append(retData, item)
					}
				} else {
					if value != filterObj.Value {
						retData = append(retData, item)
					}
				}
			} else {
				retData = append(retData, item)
			}

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
				if filterObj != nil {
					value := fmt.Sprint(item[filterObj.KeyName])
					if filterObj.Include {
						if value == filterObj.Value {
							retData = append(retData, item)
						}
					} else {
						if value != filterObj.Value {
							retData = append(retData, item)
						}
					}
				} else {
					retData = append(retData, item)
				}
			}
		} else {
			item := make(map[string]interface{})
			err := json.Unmarshal(posts[i], &item)

			if err != nil {
				logger.Debug("Error unmarshal search result json into", t, err, posts[i])
				continue
			}
			if filterObj != nil {
				value := fmt.Sprint(item[filterObj.KeyName])
				if filterObj.Include {
					if value == filterObj.Value { // note: here is compare string
						retData = append(retData, item)
					}
				} else {
					if value != filterObj.Value {
						retData = append(retData, item)
					}
				}
			} else {
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
	logger.Infof("User Search %s content Finished", search)
	//returnStructData(w, r, retData, meta)
	j, _ := json.Marshal(map[string]interface{}{
		"retCode": 0,
		"message": "ok",
		"data":    retData,
		"meta":    meta,
	})
	sendData(w, r, j)
}

// original search
func searchContent(res http.ResponseWriter, req *http.Request) {

	qs := req.URL.Query()

	t := qs.Get("type")
	// type must be set, future version may compile multi-type result set
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	it, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if hide(res, req, it()) {
		return
	}

	q, err := url.QueryUnescape(qs.Get("q"))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// q must be set
	if q == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	count, err := strconv.Atoi(qs.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if qs.Get("count") == "" {
			count = 10
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	offset, err := strconv.Atoi(qs.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if qs.Get("offset") == "" {
			offset = 0
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// execute search for query provided, if no index for type send 404
	matches, err := search.TypeQuery(t, q, count, offset)
	if err == search.ErrNoIndex {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error("[search] Error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// respond with json formatted results
	bb, err := db.ContentMulti(matches)
	if err != nil {
		logger.Error("[search] Error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// if we have matches, push the first as its matched by relevance
	if len(bb) > 0 {
		push(res, req, it(), bb[0])
	}

	var result = []json.RawMessage{}
	for i := range bb {
		result = append(result, bb[i])
	}

	j, err := fmtJSON(result...)
	if err != nil {
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
