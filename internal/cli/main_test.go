package cli

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Bool(updateFlag, false, "rewrite the command reference pages ("+referenceDir+") and the man pages ("+manDir+")")
	os.Exit(m.Run())
}
