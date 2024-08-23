package tick

import "time"

// Round rounds a duration with a precision of 3 digits
// if it is less than 100s.
// https://stackoverflow.com/a/68870075
func Round(d time.Duration) time.Duration {
	scale := 100 * time.Second
	// look for the max scale that is smaller than d
	for scale > d {
		scale = scale / 10
	}
	return d.Round(scale / 100)
}
