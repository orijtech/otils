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

		// Dereference that pointer
		if fieldVal.Kind() == reflect.Ptr {
			fieldVal = reflect.Indirect(fieldVal)
		}

		if fieldVal.Kind() == reflect.Invalid {
			continue
		}

		fieldTyp := typ.Field(i)
		if unexportedField(fieldTyp.Name) {
			continue
		}

		parentTag, omitempty, ignore := jsonTag(fieldTyp)
		if ignore {
			continue
		}

		switch fieldVal.Kind() {
		case reflect.Map:
			keys := fieldVal.MapKeys()

			for _, key := range keys {
				value := fieldVal.MapIndex(key)
				vIface := value.Interface()
				innerValueMap, err := ToURLValues(vIface)
				if err == nil && innerValueMap == nil {
					zeroValue := reflect.Zero(value.Type())
					blank := isBlank(vIface) || isBlankReflectValue(value) || reflect.DeepEqual(zeroValue.Interface(), vIface)

					if omitempty && blank {
						continue
					}
					if !blank {
						keyname := strings.Join([]string{parentTag, fmt.Sprintf("%v", key)}, ".")
						fullMap.Add(keyname, fmt.Sprintf("%v", vIface))
					}
					continue
				}

				for key, innerValueList := range innerValueMap {
					keyname := strings.Join([]string{parentTag, key}, ".")
					fullMap[keyname] = append(fullMap[keyname], innerValueList...)
				}
			}

		case reflect.Struct:
			n := fieldVal.NumField()
			typ := fieldVal.Type()
			for i := 0; i < n; i++ {
				ffield := fieldVal.Field(i)
				fTyp := typ.Field(i)
				if unexportedField(fTyp.Name) {
					continue
				}
				tag, omitempty, ignore := jsonTag(fTyp)
				if ignore {
					continue
				}
				keyname := strings.Join([]string{parentTag, tag}, ".")
				fIface := ffield.Interface()
				innerValueMap, err := ToURLValues(fIface)
				if err == nil && innerValueMap == nil {
					zeroValue := reflect.Zero(ffield.Type())
					blank := isBlank(fIface) || isBlankReflectValue(ffield) || reflect.DeepEqual(zeroValue.Interface(), fIface)
					if omitempty && blank {
						continue
					}
					if !blank {
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
			zeroValue := reflect.Zero(fieldVal.Type())
			blank := isBlank(aIface) || isBlankReflectValue(fieldVal) || reflect.DeepEqual(zeroValue.Interface(), aIface)
			if !blank {
				keyname := parentTag
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
	if val.Len() < 1 {
		return nil, nil
	}

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
			if !isBlankReflectValue(value) && !isBlank(vIface) {
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
	case "", nil, false:
		return true
	default:
		return false
	}
}

func isBlankReflectValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return v.Len() < 1
	default:
		return false
	}
}

var errInvalidValue = errors.New("invalid value")

func jsonTag(v reflect.StructField) (tag string, omitempty, ignore bool) {
	tag = v.Tag.Get("json")
	if tag == "" {
		return v.Name, false, false
	}

	splits := strings.Split(tag, ",")
	if len(splits) == 0 {
		return "", false, false
	}
	tag, instrs := splits[0], splits[1:]
	instrIndex := make(map[string]bool)
	for _, instr := range instrs {
		instrIndex[instr] = true
	}

	_, omitempty = instrIndex["omitempty"]
	_, ignore = instrIndex["-"]
	return tag, omitempty, ignore || tag == "-"
}
