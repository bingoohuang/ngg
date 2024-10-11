package dog

import "testing"

func TestRemoveFile(t *testing.T) {
	removeFiles(".", "Dog.*.pprof")
}
