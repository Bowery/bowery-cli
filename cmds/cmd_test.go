// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"testing"
)

// version.go
func TestVersionOutOfDate(t *testing.T) {
	if isOutOfDate := VersionOutOfDate("2.0.0", "2.0.1"); !isOutOfDate {
		t.Error("VersioutOutOfDate failed.")
	}
}
