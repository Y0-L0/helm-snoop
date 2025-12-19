package snooper

// Result holds the outcome of an analysis run.
// It is intended to be serialized or formatted by multiple output backends.
type Result struct {
	Referenced     []string
	DefinedNotUsed []string
	UsedNotDefined []string
}
