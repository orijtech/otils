package otils

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNullableString(t *testing.T) {
	type testNullableString struct {
		S NullableString
	}
	tests := []struct {
		name     string
		s        string
		expected NullableString
	}{
		{"normal", `{"s": "foo"}`, NullableString("foo")},
		{"null", `{"s": null}`, NullableString("")},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tn := testNullableString{S: "init"}
			if err := json.Unmarshal([]byte(tc.s), &tn); err != nil {
				t.Error(err)
			}
			if tn.S != tc.expected {
				t.Errorf("unexpected result, want: %q, got: %q", tc.expected, tn.S)
			}
		})
	}
}

func TestNullableTime(t *testing.T) {
	type testNullableTime struct {
		T *NullableTime
	}
	now := NullableTime(time.Now())
	tests := []struct {
		name     string
		s        string
		isNull   bool
		expected NullableTime
	}{
		{"null", `{"T": null}`, true, NullableTime(now)},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tn := testNullableTime{T: &now}
			if err := json.Unmarshal([]byte(tc.s), &tn); err != nil {
				t.Error(err)
			}
			if tc.isNull && tn.T != nil {
				t.Errorf("expected nil, got non-nil")
			}
			if !tc.isNull && time.Time(*tn.T).UnixNano() != time.Time(tc.expected).UnixNano() {
				t.Errorf("unexpected result, want: %v, got: %v", tc.expected, *tn.T)
			}
		})
	}
}
