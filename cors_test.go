package otils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCORSHeader(t *testing.T) {
	tests := []struct {
		name string
		cors *CORS
		want http.Header
	}{
		{
			name: "with allowCredentials",
			cors: &CORS{
				AllowCredentials: true,
			},
			want: http.Header{
				"Access-Control-Allow-Credentials": {"true"},
			},
		},
		{
			name: "all enabled",
			cors: CORSMiddlewareAllInclusive(nil).(*CORS),
			want: http.Header{
				"Access-Control-Allow-Credentials": {"true"},
				"Access-Control-Allow-Headers":     {"*"},
				"Access-Control-Allow-Methods":     {"*"},
				"Access-Control-Allow-Origin":      {"*"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.cors.setCORSForResponseWriter(rec)
			res := rec.Result()
			got := res.Header
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Mismatched end headers\nGot:  %s\nWant: %s", asJSON(got), asJSON(tt.want))
			}
		})
	}
}

func asJSON(v interface{}) []byte {
	blob, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return blob
}
