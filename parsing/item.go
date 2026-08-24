package parsing

import "fmt"

type Item struct {
	ID          string   `yaml:"-"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Obtainable  bool     `yaml:"obtainable"`
	Critical    *float64 `yaml:"critical,omitempty"`
}

func validateItems(wf WorldFile) []error {
	var errs []error

	seenItemNames := make(map[string]string) // display name -> item id
	for id, item := range wf.Items {
		if isBlank(id) {
			errs = append(errs, fmt.Errorf("an item has an empty id"))
		}
		if isBlank(item.Name) {
			errs = append(errs, fmt.Errorf("item %q: name is empty", id))
		} else if otherID, ok := seenItemNames[item.Name]; ok {
			errs = append(errs, fmt.Errorf("items %q and %q share the same name %q", otherID, id, item.Name))
		} else {
			seenItemNames[item.Name] = id
		}
		if isBlank(item.Description) {
			errs = append(errs, fmt.Errorf("item %q: description is empty", id))
		}
		if item.Critical != nil && (*item.Critical < 0 || *item.Critical > 1) {
			errs = append(errs, fmt.Errorf("item %q: critical %.2f must be between 0 and 1", id, *item.Critical))
		}
	}

	return errs
}
