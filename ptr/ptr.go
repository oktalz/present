package ptr

// New returns a pointer to the value
func New[T any](value T) *T {
	return &value
}

// Deref returns the value of the pointer
// or the zero value if the pointer is nil
func Deref[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}
