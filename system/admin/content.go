package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/item"
	"github.com/gorilla/schema"

	"time"
)

// CreateContent to create specified content
func CreateContent(ctype string, content []byte) (int, bool) {
	if content == nil {

		return 0, false
	}

	//ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))

	pt := ctype

	p, ok := item.Types[pt]
	if !ok {
		logger.Debug("no such content")
		return 0, false
	}

	post := p()
	updatecontent := map[string]interface{}{}

	err := json.Unmarshal(content, &updatecontent)
	if err != nil {
		logger.Debug("The date conversion is no ok,anyway will try to save data ", err)
	}
	ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))

	updatecontent["updated"] = ts

	upp := formatData(updatecontent)

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, upp)

	id, err := db.SetContent(ctype+":-1", upp)
	if err != nil {
		logger.Error(err.Error())

		return -1, false
	}

	return id, true
}

// UpdateContent api for update a content with key and value
func UpdateContent(t string, cid, key string, content []byte) (int, error) {

	logger.Debugf("Admin Update content %s!", t)

	pt := t

	p, ok := item.Types[pt]
	if !ok {
		logger.Debugf("Type", t, "is not a content type. Cannot edit or save.")

		return 0, errors.New("Wrong type")
	}
	ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))

	//updatecontent["updated"] = ts
	value := map[string]interface{}{}
	value[key] = string(content[:])
	value["update"] = ts
	post := p()

	upp := formatData(value)

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, upp)
	if err != nil {
		logger.Debug("Error decoding post form for edit handler:", t, err)

		return 0, errors.New("Data format issues")
	}

	id, err := db.UpdateContent(t+":"+cid, upp)
	if err != nil {
		logger.Error(err.Error())

		return -1, err
	}
	return id, nil

}

// FindContentID find a content id by key:value parameter
func FindContentID(t string, searchTxt string, searchKey string) string {

	if len(searchKey) <= 0 {
		logger.Error("search key is empty ")
		return ""
	}
	qo := db.QueryOptions{
		Count:  1,
		Offset: 1,
		Order:  "desc",
	}
	logger.Debugf("get type id with search critirial key:%s,value:%s", searchKey, searchTxt)
	total, data := db.QueryByFieldValue("Order", searchKey, searchTxt, qo)
	logger.Debugf("Total %d reocrd in %s", total, t)
	if len(data) >= 1 {
		logger.Debug(string(data[0]))
		mm := map[string]interface{}{}
		err := json.Unmarshal(data[0], &mm)
		if err != nil {
			logger.Debug("Error:", err)
			return ""
		}
		d := fmt.Sprint(mm["id"])
		return d

	}
	return ""

}
