package api

import (
	"fmt"
	"net/http"

	"github.com/agreyfox/eshop/system/item"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func omit(res http.ResponseWriter, req *http.Request, it interface{}, data []byte) ([]byte, error) {
	// is it Omittable
	om, ok := it.(item.Omittable)
	if !ok {
		return data, nil
	}

	return omitFields(res, req, om, data, "data")
}

// omit some field in map way
func omitUserFields(res http.ResponseWriter, req *http.Request, it interface{}, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// is it Omittable
	om, ok := it.(item.Omittable)
	if !ok {
		return data, nil
	}
	fields, err := om.Omit(res, req)
	if err != nil {
		logger.Debug(err)
		return data, err
	}

	for i := 0; i < len(data); i++ {
		item := data[i]
		for j := 0; j < len(fields); j++ {
			delete(item, fields[j])
		}
	}

	return data, nil
}

func omitFields(res http.ResponseWriter, req *http.Request, om item.Omittable, data []byte, pathPrefix string) ([]byte, error) {
	// get fields to omit from json data
	fields, err := om.Omit(res, req)
	if err != nil {
		return nil, err
	}

	// remove each field from json, all responses contain json object(s) in top-level "data" array
	n := int(gjson.GetBytes(data, pathPrefix+".#").Int())
	for i := 0; i < n; i++ {
		for k := range fields {
			var err error
			data, err = sjson.DeleteBytes(data, fmt.Sprintf("%s.%d.%s", pathPrefix, i, fields[k]))
			if err != nil {
				logger.Error("Error omitting field:", fields[k], "from item.Omittable:", om)
				return nil, err
			}
		}
	}

	return data, nil
}
