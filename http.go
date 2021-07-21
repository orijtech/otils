package otils

import (
	"fmt"
	"net/http"
)

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
