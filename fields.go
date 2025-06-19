package labeler

// FieldFlag represents a bitmask for fields that can be evaluated for labeling.
// Use bitwise operations or helper methods to combine and check flags.
type FieldFlag uint

// Has returns true if the FieldFlag contains the provided flag.
func (f FieldFlag) Has(flag FieldFlag) bool {
	return f&flag != 0
}

// OrDefault returns AllFieldFlags if no flags are set (f == 0), otherwise returns f.
func (f FieldFlag) OrDefault() FieldFlag {
	if f == 0 {
		return AllFieldFlags
	}
	return f
}

const (
	// FieldTitle indicates the title field should be evaluated for labeling.
	FieldTitle FieldFlag = 1 << iota
	// FieldBody indicates the body field should be evaluated for labeling.
	FieldBody

	// AllFieldFlags is a convenience constant representing all available fields.
	AllFieldFlags = FieldTitle | FieldBody
)

// ParseFieldFlags converts a slice of string field names to a FieldFlag bitmask.
// Unrecognized field names are ignored.
func ParseFieldFlags(fields []string) FieldFlag {
	var flags FieldFlag
	for _, f := range fields {
		switch f {
		case "title":
			flags |= FieldTitle
		case "body":
			flags |= FieldBody
		}
	}
	return flags
}
