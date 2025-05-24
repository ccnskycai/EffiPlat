package model

// StringPtr returns a pointer to the string value s.
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the int value i.
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the bool value b.
func BoolPtr(b bool) *bool {
	return &b
}
