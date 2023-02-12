package labeler

// ptr creates a pointer to a value, avoiding need for extracting to an assignment line first
func ptr[T any](value T) *T {
	return &value
}
