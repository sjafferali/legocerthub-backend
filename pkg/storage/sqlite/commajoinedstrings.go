package sqlite

import (
	"strings"
)

// commaJoinedStrings is a string type in storage that is a list of
// strings joined by commas
type commaJoinedStrings string

// transform CJS into string slice
func (cjs commaJoinedStrings) toSlice() []string {
	if cjs == "" {
		return []string{}
	}

	return strings.Split(string(cjs), ",")
}

// makeCommaJoinedString creates a CJS from a slice of strings
func makeCommaJoinedString(stringSlice []string) commaJoinedStrings {
	if len(stringSlice) == 0 {
		return ""
	}

	return commaJoinedStrings(strings.Join(stringSlice, ","))
}
