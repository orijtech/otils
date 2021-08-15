package otils

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

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
		*ns = ""
		return nil
	}
	unquoted, err := strconv.Unquote(str)
	if err != nil {
		return err
	}
	*ns = NullableString(unquoted)

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
