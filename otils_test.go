package otils_test

import (
	"encoding/json"
	"testing"

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
			want: "logo.dimension.extra.shade=45%25&logo.dimension.extra.zoom=false&logo.dimension.height=120&logo.dimension.width=100&logo.url=https%3A%2F%2Forijtech.com%2Ffavicon.ico&source=https%3A%2F%2Forijtech.com",
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
						Extra: map[string]interface{}{
							"zoom": false, "shade": "0%",
						},
					},
				},
			},
			want: "logo.dimension.extra.shade=0%25&logo.dimension.extra.zoom=false&logo.dimension.height=120&logo.dimension.width=100&logo.url=https%3A%2F%2Forijtech.com%2Ffavicon.ico",
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
	Width  int `json:"width"`
	Height int `json:"height"`

	Extra map[string]interface{} `json:"extra,omitempty"`
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
