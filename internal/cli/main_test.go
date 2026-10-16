package cli

import (
	"flag"
	"os"
	"testing"
)

// TestMain registers the flags of the package's tests, then runs them:
// -update makes TestCommandReference rewrite the command reference pages.
func TestMain(m *testing.M) {
	flag.Bool(updateFlag, false, "rewrite the command reference pages ("+referenceDir+")")
	os.Exit(m.Run())
}
