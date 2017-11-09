package otils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
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

// RedirectAllTrafficTo creates a handler that can be attached
// to an HTTP traffic multiplexer to perform a 301 Permanent Redirect
// to the specified host for any path, anytime that the handler
// receives a request.
// Sample usage is:
//
//  httpsRedirectHandler := RedirectAllTrafficTo("https://orijtech.com")
//  if err := http.ListenAndServe(":80", httpsRedirectHandler); err != nil {
//    log.Fatal(err)
//  }
//
// which is used in production at orijtech.com to redirect any non-https
// traffic from http://orijtech.com/* to https://orijtech.com/*
func RedirectAllTrafficTo(host string) http.Handler {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		finalURL := fmt.Sprintf("%s%s", host, req.URL.Path)
		rw.Header().Set("Location", finalURL)
		rw.WriteHeader(301)
	}

	return http.HandlerFunc(fn)
}

// StatusOK returns true if a status code is a 2XX code
func StatusOK(code int) bool { return code >= 200 && code <= 299 }

// FirstNonEmptyString iterates through its
// arguments trying to find the first string
// that is not blank or consists entirely  of spaces.
func FirstNonEmptyString(args ...string) string {
	for _, arg := range args {
		if arg == "" {
			continue
		}
		if strings.TrimSpace(arg) != "" {
			return arg
		}
	}
	return ""
}

func NonEmptyStrings(args ...string) (nonEmpties []string) {
	for _, arg := range args {
		if arg == "" {
			continue
		}
		if strings.TrimSpace(arg) != "" {
			nonEmpties = append(nonEmpties, arg)
		}
	}
	return nonEmpties
}

type CodedError struct {
	code int
	msg  string
}

func (cerr *CodedError) Error() string {
	if cerr == nil {
		return ""
	}
	return cerr.msg
}

func (cerr *CodedError) Code() int {
	if cerr == nil {
		return http.StatusOK
	}
	return cerr.code
}

func MakeCodedError(msg string, code int) *CodedError {
	return &CodedError{
		msg:  msg,
		code: code,
	}
}

// NullableString represents a string that is sent
// back by some APIs as null, in JSON unquoted which
// makes them un-unmarshalable in Go.
// NullableString interprets null as "".
type NullableString string

var _ json.Unmarshaler = (*NullableString)(nil)

func (ns *NullableString) UnmarshalJSON(b []byte) error {
	str := string(b)
	// Special case when we encounter `null`, modify it to the empty string
	if str == "null" || str == "" {
		str = ""
	} else {
		unquoted, err := strconv.Unquote(str)
		if err != nil {
			return err
		}
		*ns = NullableString(unquoted)
	}

	return nil
}

type NullableFloat64 float64

var _ json.Unmarshaler = (*NullableFloat64)(nil)

func (nf64 *NullableFloat64) UnmarshalJSON(b []byte) error {
	str := string(b)
	if strings.HasPrefix(str, "\"") {
		unquoted, err := strconv.Unquote(str)
		if err == nil {
			str = unquoted
		}
	}

	f64, err := strconv.ParseFloat(str, 64)
	if err == nil {
		*nf64 = NullableFloat64(f64)
		return nil
	}

	// Otherwise trying checking if it was null
	var ns NullableString
	if err := json.Unmarshal(b, &ns); err != nil {
		return err
	}

	if ns == "" {
		*nf64 = 0.0
		return nil
	}

	f64, err = strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}

	*nf64 = NullableFloat64(f64)
	return nil
}

type NumericBool bool

func (nb *NumericBool) UnmarshalJSON(blob []byte) error {
	if len(blob) < 1 {
		*nb = false
		return nil
	}

	s := string(blob)
	// Try first parsing an integer.
	pBool, err := strconv.ParseBool(s)
	if err == nil {
		*nb = NumericBool(pBool)
		return nil
	}

	pInt, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		*nb = pInt != 0
		return nil
	}

	return err
}

type NullableTime time.Time

var _ json.Unmarshaler = (*NullableTime)(nil)

func (nt *NullableTime) UnmarshalJSON(b []byte) error {
	var ns NullableString
	if err := json.Unmarshal(b, &ns); err != nil {
		return err
	}
	if ns == "" {
		nt = nil
		return nil
	}

	// To parse the time, we need to quote the value
	quotedStr := strconv.Quote(string(ns))
	t := new(time.Time)
	if err := json.Unmarshal([]byte(quotedStr), t); err != nil {
		return err
	}

	*nt = NullableTime(*t)
	return nil
}

type CORS struct {
	Origins []string
	Methods []string
	Headers []string

	next http.Handler
}

func CORSMiddleware(c *CORS, next http.Handler) http.Handler {
	if c == nil {
		return next
	}
	copy := new(CORS)
	*copy = *c
	copy.next = next
	return copy
}

var allInclusiveCORS = &CORS{
	Origins: []string{"*"},
	Methods: []string{"*"},
	Headers: []string{"*"},
}

// CORSMiddlewareAllInclusive is a convenience helper that uses the
// all inclusive CORS:
// Access-Control-Allow-Origin: *
// Access-Control-Allow-Methods: *
// Access-Control-Allow-Headers: *
// thus enabling all origins, all methods and all headers.
func CORSMiddlewareAllInclusive(next http.Handler) http.Handler {
	return CORSMiddleware(allInclusiveCORS, next)
}

var _ http.Handler = (*CORS)(nil)

func (c *CORS) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c.setCORSForResponseWriter(rw)
	if c.next != nil {
		c.next.ServeHTTP(rw, req)
	}
}

func (c *CORS) setCORSForResponseWriter(rw http.ResponseWriter) {
	for _, origin := range c.Origins {
		rw.Header().Add("Access-Control-Allow-Origin", origin)
	}
	for _, mtd := range c.Methods {
		rw.Header().Add("Access-Control-Allow-Methods", mtd)
	}
	for _, hdr := range c.Headers {
		rw.Header().Add("Access-Control-Allow-Headers", hdr)
	}
}

func EnvOrAlternates(envVar string, alternates ...string) string {
	if retr := strings.TrimSpace(os.Getenv(envVar)); retr != "" {
		return retr
	}
	for _, alt := range alternates {
		alt = strings.TrimSpace(alt)
		if alt != "" {
			return alt
		}
	}
	return ""
}

func unexportedField(name string) bool {
	return len(name) > 0 && name[0] >= 'a' && name[0] <= 'z'
}

func UniqStrings(strs ...string) []string {
	uniqs := make([]string, 0, len(strs))
	seen := make(map[string]bool)
	for _, str := range strs {
		if _, ok := seen[str]; !ok {
			seen[str] = true
			uniqs = append(uniqs, str)
		}
	}
	return uniqs
}
