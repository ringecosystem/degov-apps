package utils

import "time"

// TimePtr returns a pointer to the given time value.
// This is useful when you need to pass a time pointer to a struct field.
func TimePtr(t time.Time) *time.Time {
	return &t
}

func TimePtrNow() *time.Time {
	t := time.Now()
	return &t
}

// StringPtr returns a pointer to the given string value.
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int value.
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr returns a pointer to the given int64 value.
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// Float64Ptr returns a pointer to the given float64 value.
func Float64Ptr(f float64) *float64 {
	return &f
}
