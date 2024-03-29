package otils_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/orijtech/otils"
)

func TestToURLValues(t *testing.T) {
	tests := [...]struct {
		v       interface{}
		want    string
		mustErr bool
	}{
		0: {
			v: &Request{
				Source: "https://orijtech.com",
				Logo: &Logo{
					URL: "https://orijtech.com/favicon.ico",
					Dimensions: &Dimension{
						Width: 100, Height: 120,
						Extra: map[string]interface{}{
							"zoom": false, "shade": "45%",
						},
					},
				},
			},
			want: "logo.dimension.extra.shade=45%25&logo.dimension.height=120&logo.dimension.width=100&logo.url=https%3A%2F%2Forijtech.com%2Ffavicon.ico&source=https%3A%2F%2Forijtech.com",
		},

		1: {
			v:       nil,
			mustErr: true,
		},

		2: {
			v:       "thisway",
			mustErr: true,
		},

		3: {
			v: Request{
				Logo: &Logo{
					URL: "https://orijtech.com/favicon.ico",
					Dimensions: &Dimension{
						Width: 100, Height: 120,
						BasicName: "flux",
						Inner:     &Dimension{Width: 120, Height: 240},
						Extra: map[string]interface{}{
							"zoom": false, "shade": "0%",
						},
					},
				},
			},
			want: "logo.dimension.extra.shade=0%25&logo.dimension.height=120&logo.dimension.width=100&logo.url=https%3A%2F%2Forijtech.com%2Ffavicon.ico",
		},

		4: {
			v: []*Request{
				{Logo: &Logo{URL: "https://orijtech.com/favicon.ico"}},
				nil,
			},
			want: "0=logo.url%3Dhttps%253A%252F%252Forijtech.com%252Ffavicon.ico",
		},

		5: {
			v:    map[string]int{"uno": 1, "zero": 0, "satu": 3, "saba": 7},
			want: "saba=7&satu=3&uno=1&zero=0",
		},

		6: {
			v:       func() int { return 2 },
			mustErr: true,
		},

		7: {
			v:    &Query{Nested: true, Page: 0},
			want: "nested=true",
		},

		8: {
			v:    &Query{Nested: false, Page: 0},
			want: "",
		},

		9: {
			v:    &Query{Nested: false, Page: 2},
			want: "page=2",
		},
	}

	for i, tt := range tests {
		values, err := otils.ToURLValues(tt.v)
		if tt.mustErr {
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}
		got, want := values.Encode(), tt.want
		if got != want {
			t.Errorf("#%d:\ngot:  %q\nwant: %q", i, got, want)
		}
	}
}

type Dimension struct {
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	BasicName string     `json:"-"`
	Inner     *Dimension `json:"-"`

	Extra map[string]interface{} `json:"extra,omitempty"`
}

type Query struct {
	Nested bool  `json:"nested"`
	Page   int64 `json:"page"`
}

type Logo struct {
	URL        string     `json:"url"`
	Dimensions *Dimension `json:"dimension"`
}

type Request struct {
	Logo   *Logo  `json:"logo"`
	Source string `json:"source"`
}

func TestFirstNonEmptyString(t *testing.T) {
	tests := [...]struct {
		args []string
		want string
	}{
		0: {args: []string{"     ", "", "a", "b"}, want: "a"},
		1: {args: []string{""}, want: ""},
		2: {args: []string{"", " "}, want: ""},
		3: {args: []string{"ABC", "DEF", " "}, want: "ABC"},
		4: {args: []string{"", " DEF ", " "}, want: " DEF "},
	}

	for i, tt := range tests {
		got := otils.FirstNonEmptyString(tt.args...)
		want := tt.want
		if got != want {
			t.Errorf("#%d got=%q want=%q", i, got, want)
		}
	}
}

func TestCodedError(t *testing.T) {
	// No panics expected
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %#v", r)
		}
	}()

	tests := [...]struct {
		err  *otils.CodedError
		msg  string
		code int
	}{
		0: {otils.MakeCodedError("200 OK", 200), "200 OK", 200},
		1: {otils.MakeCodedError("failed to find it", 404), "failed to find it", 404},

		// nil CodedError should not panic and should return 200 values.
		2: {nil, "", 200},
	}

	for i, tt := range tests {
		gotCode, wantCode := tt.err.Code(), tt.code
		if gotCode != wantCode {
			t.Errorf("#%d gotCode=%v wantCode=%v", i, gotCode, wantCode)
		}

		gotMsg, wantMsg := tt.err.Error(), tt.msg
		if gotMsg != wantMsg {
			t.Errorf("#%d gotMsg=%v wantMsg=%v", i, gotMsg, wantMsg)
		}
	}
}

func TestNumericBool(t *testing.T) {
	tests := [...]struct {
		str     string
		want    otils.NumericBool
		wantErr bool
	}{
		0: {str: "1", want: otils.NumericBool(true)},
		1: {str: "0", want: otils.NumericBool(false)},
		2: {str: "true", want: otils.NumericBool(true)},
		3: {str: "false", want: otils.NumericBool(false)},
		4: {str: "ping", wantErr: true},
	}

	for i, tt := range tests {
		var nb otils.NumericBool
		err := json.Unmarshal([]byte(tt.str), &nb)

		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d: expecting non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}

		got, want := nb, tt.want
		if got != want {
			t.Errorf("#%d got=%v want=%v", i, got, want)
		}
	}
}

func TestNullableTime(t *testing.T) {
	tests := [...]struct {
		str     string
		want    otils.NullableTime
		wantErr bool
	}{
		0: {str: "", wantErr: true},
		1: {str: "0", wantErr: true},
		2: {
			str:  `"2006-01-02T15:04:05Z"`,
			want: otils.NullableTime(time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC)),
		},
		3: {str: "ping", wantErr: true},
	}

	for i, tt := range tests {
		var nt otils.NullableTime
		err := json.Unmarshal([]byte(tt.str), &nt)

		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d: expecting non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err: %v", i, err)
			continue
		}

		tGot, tWant := time.Time(nt), time.Time(tt.want)
		if !tGot.Equal(tWant) {
			t.Errorf("#%d got=%v want=%v", i, tGot, tWant)
		}
	}
}

func TestNullableFloat64(t *testing.T) {
	tests := [...]struct {
		str     string
		wantErr bool
		want    otils.NullableFloat64
	}{
		0: {
			str: "", wantErr: true,
		},
		1: {
			str: `"-5.7"`, want: -5.7,
		},
		2: {
			str: `"00010.9"`, want: 10.9,
		},
		3: {
			str: `null`, want: 0,
		},
	}

	for i, tt := range tests {
		var recv otils.NullableFloat64
		err := json.Unmarshal([]byte(tt.str), &recv)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d: expected a non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: gotErr=%v", i, err)
			continue
		}

		// Compare by JSON repr
		gotBytes, wantBytes := jsonify(recv), jsonify(tt.want)
		if !bytes.Equal(gotBytes, wantBytes) {
			t.Errorf("#%d:\ngot: %s\nwant:%s", i, gotBytes, wantBytes)
		}
	}
}

func jsonify(v interface{}) []byte {
	blob, _ := json.MarshalIndent(v, "", "  ")
	return blob
}

func TestCORSMiddleware(t *testing.T) {
	tests := [...]struct {
		cors *otils.CORS
		want http.Header
	}{
		0: {want: nil},
		1: {
			cors: &otils.CORS{
				Headers: []string{"X-Preflight"},
				Methods: []string{"POST", "GET"},
			}, want: http.Header{
				"Access-Control-Allow-Methods": []string{"POST", "GET"},
				"Access-Control-Allow-Headers": []string{"X-Preflight"},
			},
		},
		2: {
			cors: &otils.CORS{
				Methods: []string{"POST", "DELETE"},
			}, want: http.Header{
				"Access-Control-Allow-Methods": []string{"POST", "DELETE"},
			},
		},
	}

	for i, tt := range tests {
		handler := otils.CORSMiddleware(tt.cors, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("Hello!"))
		}))
		tst := httptest.NewServer(handler)
		res, err := tst.Client().Get(tst.URL)
		tst.Close()
		if err != nil {
			t.Errorf("#%d unexpected err: %v", i, err)
			continue
		}
		// Match headers
		for wantKey, wantValues := range tt.want {
			gotValues := res.Header[wantKey]
			if !reflect.DeepEqual(gotValues, wantValues) {
				t.Errorf("#%d: key: %v\ngotHeader: %#v\nwantHeader:%#v", i, wantKey, gotValues, wantValues)
			}
		}
	}
}
