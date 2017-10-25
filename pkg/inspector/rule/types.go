package rule

// Meta contains the rule's metadata
type Meta struct {
	Kind string
	When [][]string
}

// GetRuleMeta returns the rule's metadata
func (rm Meta) GetRuleMeta() Meta {
	return rm
}

// Rule is an inspector rule
type Rule interface {
	Name() string
	GetRuleMeta() Meta
	IsRemoteRule() bool
	Validate() []error
}

// Result contains the results from executing the rule
type Result struct {
	// Name is the rule's name
	Name string
	// Success is true when the rule was asserted
	Success bool
	// Error message if there was an error executing the rule
	Error string
	// Remediation contains potential remediation steps for the rule
	Remediation string
}
