package models

import "time"

// -- Utility function

// TimePtrToStringPtr uses TimeToString to optonally format a string
func TimePtrToStringPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := TimeToString(*t)
	return &s
}

// TimeToString converts a time to UTC, then formats as RFC3339
func TimeToString(t time.Time) string {

	return t.Round(0).UTC().Format(time.RFC3339)
}

// StringToTime converts a RFC3339 formatted string into a time.Time
func StringToTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return t, err
	}
	return t.Round(0), nil
}

// StringToTimeMust works like StringToTime but panics on errors.
// I think this is usually acceptable as times are formatted pretty carefully
// in the db
func StringToTimeMust(s string) time.Time {
	t, err := StringToTime(s)
	if err != nil {
		panic(err)
	}
	return t.Round(0)
}

// BoolToInt64 converts a bool to int64 (1 for true, 0 for false)
func BoolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Int64ToBool converts an int64 to bool (non-zero is true)
func Int64ToBool(i int64) bool {
	return i != 0
}

// BoolPtrToInt64Ptr converts a *bool to *int64
func BoolPtrToInt64Ptr(b *bool) *int64 {
	if b == nil {
		return nil
	}
	i := BoolToInt64(*b)
	return &i
}
