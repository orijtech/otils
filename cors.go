package otils

import (
	"net/http"
)

type CORS struct {
	Origins []string
	Methods []string
	Headers []string

	// AllowCredentials when set signifies that the header
	// "Access-Control-Allow-Credentials" which will allow
	// the possibility of the frontend XHR's withCredentials=true
	// to be set.
	AllowCredentials bool

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

	AllowCredentials: true,
}

// CORSMiddlewareAllInclusive is a convenience helper that uses the
// all inclusive CORS:
// Access-Control-Allow-Origin: *
// Access-Control-Allow-Methods: *
// Access-Control-Allow-Headers: *
// Access-Control-Allow-Credentials: *
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
	if c.AllowCredentials {
		rw.Header().Add("Access-Control-Allow-Credentials", "true")
	}
}

func unexportedField(name string) bool {
	return len(name) > 0 && name[0] >= 'a' && name[0] <= 'z'
}
