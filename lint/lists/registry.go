package lists

var listRegistry = make(map[string][]string)

// Register adds a named list of rule IDs to the registry.
func Register(name string, ruleIDs []string) {
	listRegistry[name] = ruleIDs
}

// Get returns the rule IDs for a given list name, and a boolean indicating if it was found.
func Get(name string) ([]string, bool) {
	ids, ok := listRegistry[name]
	return ids, ok
}

// GetAll returns all registered lists
func GetAll() map[string][]string {
	// Return a copy to prevent modification
	ret := make(map[string][]string)
	for k, v := range listRegistry {
		ret[k] = v
	}
	return ret
}
