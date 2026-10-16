package cli

import (
	"context"
	"os"
	"testing"
)

// TestIsTerminal checks that neither a file, open or closed, nor the null
// device, which scripts give as input (and Windows reports as the same file
// as a console), nor a pipe is a terminal; nor can ReadHidden read one.
func TestIsTerminal(t *testing.T) {
	file, err := os.Open(writeBody(t, "{}"))
	if err != nil {
		t.Fatal(err)
	}
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = null.Close() })
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
	if IsTerminal(file) || IsTerminal(null) || IsTerminal(reader) {
		t.Fatal("a file, the null device or a pipe is a terminal")
	}
	if _, err := ReadHidden(null)(context.Background()); err == nil {
		t.Error("ReadHidden read the null device")
	}
	_ = file.Close()
	if IsTerminal(file) {
		t.Fatal("a closed file is a terminal")
	}
}
