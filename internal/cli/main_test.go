package cli

import (
	"flag"
	"os"
	"testing"
)

// TestMain registers the flags of the package's tests, then runs them:
// -update makes TestCommandReference and TestManPages rewrite the command
// reference pages and the man pages.
func TestMain(m *testing.M) {
	flag.Bool(updateFlag, false, "rewrite the command reference pages ("+referenceDir+") and the man pages ("+manDir+")")
	os.Exit(m.Run())
}
