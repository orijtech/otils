package otils

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// ToURLValues transforms any type with fields into a url.Values map
// so that it can be used to make the QUERY string in HTTP GET requests
// for example:
//
// Transforming a struct whose JSON representation is:
// {
//    "url": "https://orijtech.com",
//    "logo": {
//	"url": "https://orijtech.com/favicon.ico",
//	"dimens": {
//	  "width": 100, "height": 120,
//	  "extra": {
//	    "overlap": true,
//	    "shade": "48%"
//	  }
//	}
//    }
// }
//
// Into:
// "logo.dimension.extra.shade=48%25&logo.dimension.extra.zoom=false&logo.dimension.height=120&logo.dimension.width=100&logo.url=https%3A%2F%2Forijtech.com%2Ffavicon.ico"
func ToURLValues(v interface{}) (url.Values, error) {
	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.Ptr:
		val = reflect.Indirect(val)
	case reflect.Struct:
		// Let this pass through
	case reflect.Array, reflect.Slice:
		return toURLValuesForSlice(v)
	case reflect.Map:
		return toURLValuesForMap(v)
	default:
		return nil, nil
	}

	fullMap := make(url.Values)
	if !val.IsValid() {
		return nil, errInvalidValue
	}

	typ := val.Type()
	nfields := val.NumField()

	for i := 0; i < nfields; i++ {
		fieldVal := val.Field(i)
		if fieldVal.Kind() == reflect.Ptr {
			fieldVal = reflect.Indirect(fieldVal)
		}

		fieldTyp := typ.Field(i)
		jsonTag := firstJSONTag(fieldTyp)

		switch fieldVal.Kind() {
		case reflect.Map:
			keys := fieldVal.MapKeys()

			for _, key := range keys {
				value := fieldVal.MapIndex(key)
				vIface := value.Interface()
				innerValueMap, err := ToURLValues(vIface)
				if err == nil && innerValueMap == nil {
					if !isBlank(vIface) {
						keyname := strings.Join([]string{jsonTag, fmt.Sprintf("%v", key)}, ".")
						fullMap.Add(keyname, fmt.Sprintf("%v", vIface))
					}
					continue
				}

				for key, innerValueList := range innerValueMap {
					keyname := strings.Join([]string{jsonTag, key}, ".")
					fullMap[keyname] = append(fullMap[keyname], innerValueList...)
				}
			}

		case reflect.Struct:
			n := fieldVal.NumField()
			typ := fieldVal.Type()
			for i := 0; i < n; i++ {
				ffield := fieldVal.Field(i)
				fTyp := typ.Field(i)
				keyname := strings.Join([]string{jsonTag, firstJSONTag(fTyp)}, ".")

				fIface := ffield.Interface()
				innerValueMap, err := ToURLValues(fIface)
				if err == nil && innerValueMap == nil {
					if !isBlank(fIface) {
						fullMap.Add(keyname, fmt.Sprintf("%v", fIface))
					}
					continue
				}

				for key, innerValueList := range innerValueMap {
					keyname := strings.Join([]string{keyname, key}, ".")
					fullMap[keyname] = append(fullMap[keyname], innerValueList...)
				}
			}

		default:
			aIface := fieldVal.Interface()
			if !isBlank(aIface) {
				keyname := jsonTag
				fullMap[keyname] = append(fullMap[keyname], fmt.Sprintf("%v", aIface))
			}
		}
	}

	return fullMap, nil
}

func toURLValuesForSlice(v interface{}) (url.Values, error) {
	val := reflect.ValueOf(v)
	n := val.Len()
	finalValues := make(url.Values)

	sliceValues := val.Slice(0, val.Len())
	for i := 0; i < n; i++ {
		ithVal := sliceValues.Index(i)
		iface := ithVal.Interface()
		// Goal here is to recombine them into
		// {0: url.Values}
		retr, _ := ToURLValues(iface)
		if len(retr) > 0 {
			key := fmt.Sprintf("%d", i)
			finalValues[key] = append(finalValues[key], retr.Encode())
		}
	}

	return finalValues, nil
}

func toURLValuesForMap(v interface{}) (url.Values, error) {
	val := reflect.ValueOf(v)
	keys := val.MapKeys()

	fullMap := make(url.Values)
	for _, key := range keys {
		value := val.MapIndex(key)
		vIface := value.Interface()
		keyname := fmt.Sprintf("%v", key)
		innerValueMap, err := ToURLValues(vIface)
		if err == nil && innerValueMap == nil {
			if !isBlank(vIface) {
				fullMap.Add(keyname, fmt.Sprintf("%v", vIface))
			}
			continue
		}

		for key, innerValueList := range innerValueMap {
			innerKeyname := strings.Join([]string{keyname, key}, ".")
			fullMap[innerKeyname] = append(fullMap[innerKeyname], innerValueList...)
		}
	}

	return fullMap, nil
}

// isBlank returns true if a value will leave a value blank in a URL Query string
// e.g:
//  * `value=`
//  * `value=null`
func isBlank(v interface{}) bool {
	switch v {
	case "", nil:
		return true
	default:
		return false
	}
}

var errInvalidValue = errors.New("invalid value")

func firstJSONTag(v reflect.StructField) string {
	jsonTag := v.Tag.Get("json")
	if jsonTag == "" {
		jsonTag = v.Name
	} else {
		// In the case that we have say `json:"name,omitempty"`
		// we only want "name" as the tag, the rest
		splits := strings.Split(jsonTag, ",")
		jsonTag = splits[0]
	}
	return jsonTag
}

