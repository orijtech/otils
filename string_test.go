package otils

import (
	"reflect"
	"sort"
	"testing"
)

func TestUniqStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"ok", []string{"b", "c", "c", "a", "b"}, []string{"a", "b", "c"}},
		{"nil", nil, []string{}},
		{"empty", []string{}, []string{}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := UniqStrings(tc.input...)
			sort.Strings(got)
			sort.Strings(tc.expected)
			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf("unexpected result, want: %v, got: %v", tc.expected, got)
			}
		})
	}
}
