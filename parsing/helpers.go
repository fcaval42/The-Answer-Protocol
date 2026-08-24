package parsing

import "strings"

// isBlank reports whether s is empty once surrounding whitespace is stripped,
// so a description made of only spaces is treated the same as "".
func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// duplicatesIn returns each value of refs that appears more than once,
// ignoring empty entries (those are reported separately as missing references).
func duplicatesIn(refs []string) []string {
	seen := make(map[string]bool, len(refs))
	var dups []string
	for _, ref := range refs {
		if ref == "" {
			continue
		}
		if seen[ref] {
			dups = append(dups, ref)
		}
		seen[ref] = true
	}
	return dups
}
