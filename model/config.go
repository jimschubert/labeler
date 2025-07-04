package model

// Config is the interface used by simple and full config objects
type Config interface {
	// FromBytes is used to parse bytes into the Config instance
	FromBytes(b []byte) error

	// LabelsFor allows config implementations to determine the labels to be applied to the input strings
	LabelsFor(text ...string) map[string]Label
}

type FieldOverrides interface {
	// IncludedFields returns the fields that are used for labeling, if not defined, it returns an empty slice
	IncludedFields() []string
}
