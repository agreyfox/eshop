package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

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
	updatecontent["timestamp"] = ts //2021/3/1重要，首次创建
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

// RestoreContent to specified ctype and report error
func RestoreContent(ctype string, cid int, content []byte) error {
	if content == nil || len(content) == 0 {

		return fmt.Errorf("no content need to be restore")
	}

	//ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))

	pt := ctype

	p, ok := item.Types[pt]
	if !ok {
		logger.Errorf("no such content in the content system ")
		return fmt.Errorf("no such content in the content system")
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

	_, err = db.SetContent(ctype+":"+fmt.Sprint(cid), upp)
	if err != nil {
		logger.Error(err.Error())

		return err
	}

	return nil
}

// update ad content by muilt filed
func UpdateContents(t string, cid string, key []string, updateValue *url.Values) (int, error) {

	logger.Debugf("Admin Update content %s! with multfeild %v", t, key)

	pt := t

	p, ok := item.Types[pt]
	if !ok {
		logger.Debugf("Type", t, "is not a content type. Cannot edit or save.")

		return 0, errors.New("Wrong type")
	}
	ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))

	//updatecontent["updated"] = ts

	value := map[string]interface{}{}
	for _, item := range key {
		value[item] = string(updateValue.Get(item))
	}

	value["updated"] = ts
	post := p()

	upp := formatData(value)

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, upp)
	if err != nil {
		logger.Debug("Error decoding post form for edit handler:", t, err)
		return 0, fmt.Errorf("data format issues")
	}

	id, err := db.UpdateContent(t+":"+cid, upp)
	if err != nil {
		logger.Error(err.Error())

		return -1, err
	}
	return id, nil
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
	value["updated"] = ts
	post := p()

	upp := formatData(value)

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, upp)
	if err != nil {
		logger.Debug("Error decoding post form for edit handler:", t, err)
		return 0, fmt.Errorf("Data format issues:%s", err.Error())
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
		Offset: 0, //找第一个页
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

// search all key in list by content type

func GetAllKeyOfContent(ctype string) ([]string, error) {
	total, keylist := db.QueryContentKey(ctype, true)
	if total == 0 {
		return nil, fmt.Errorf("no key found")
	} else {
		return keylist, nil
	}

}
