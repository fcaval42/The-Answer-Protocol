package parsing

import "strings"

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

var locationNameConnectors = map[string]bool{
	"de": true, "d": true, "du": true, "des": true,
	"le": true, "la": true, "l": true,
}

func slugifyLocationName(name string) string {
	fields := strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '\'' || r == '’'
	})

	var parts []string
	for _, f := range fields {
		if locationNameConnectors[strings.ToLower(f)] {
			continue
		}
		if len(parts) == 0 {
			parts = append(parts, f)
		} else {
			parts = append(parts, strings.ToLower(f))
		}
	}

	return strings.Join(parts, "_")
}

func slugifyName(name string) string {
	replacer := strings.NewReplacer(" ", "_", "-", "_", "'", "_", "’", "_")
	return replacer.Replace(name)
}

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
