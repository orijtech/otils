package otils

import "strings"

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
