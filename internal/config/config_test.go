package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestLoadWithoutAFileHasNoProfiles(t *testing.T) {
	f, err := In(t.TempDir()).Load()
	if err != nil || f.Default != "" || f.Profiles == nil || len(f.Profiles) != 0 {
		t.Fatalf("file %+v, %v; want no profiles", f, err)
	}
}

func TestSaveThenLoadKeepsTheProfiles(t *testing.T) {
	store := In(t.TempDir())
	want := File{Default: "dev", Profiles: map[string]Profile{
		"dev":  {Tenant: "acme1", ClientID: "c1", Site: "AcmeStore_NL", Language: "nl"},
		"prod": {Tenant: "acme2", APIURL: "https://api.example.com", ClientID: "c2"},
	}}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded %+v, %v; want %+v", got, err, want)
	}
	if !strings.HasSuffix(store.Path(), filepath.Join("epx", "config.json")) {
		t.Errorf("path %s; want epx/config.json", store.Path())
	}
	if runtime.GOOS == "windows" {
		return
	}
	for path, mode := range map[string]os.FileMode{store.Path(): 0o600, filepath.Dir(store.Path()): 0o700} {
		if info, err := os.Stat(path); err != nil || info.Mode().Perm() != mode {
			t.Errorf("%s: mode %v, %v; want %v", path, info.Mode().Perm(), err, mode)
		}
	}
	if entries, _ := os.ReadDir(filepath.Dir(store.Path())); len(entries) != 1 {
		t.Errorf("directory holds %d files; want the configuration only", len(entries))
	}
}

func TestLoadRefusesAFileOthersCanAccess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows reports no access of other users in a file mode")
	}
	store := In(t.TempDir())
	if err := store.Save(File{}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []os.FileMode{0o644, 0o640, 0o602} {
		if err := os.Chmod(store.Path(), mode); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Load(); err == nil || !strings.Contains(err.Error(), "chmod 600") {
			t.Errorf("mode %v: %v; want a refusal", mode, err)
		}
	}
}

func TestLoadRefusesWhatIsNoConfiguration(t *testing.T) {
	const secret = "s3cr3t-value"
	for name, content := range map[string]string{
		"not JSON":           secret,
		"a client secret":    `{"profiles":{"dev":{"tenant":"acme1","clientId":"c","clientSecret":"` + secret + `"}}}`,
		"a wrong type":       `{"profiles":{"dev":{"tenant":["` + secret + `"]}}}`,
		"two values":         `{"profiles":{}} {"x":"` + secret + `"}`,
		"an invalid name":    `{"profiles":{"` + secret + ` x":{"tenant":"acme1"}}}`,
		"a name in capitals": `{"profiles":{"` + strings.ToUpper(secret) + `":{"tenant":"acme1"}}}`,
		"a stray default":    `{"default":"` + secret + `","profiles":{}}`,
		"an unknown root":    `{"` + secret + `":1}`,
		"a trailing brace":   `{"profiles":{}}}`,
	} {
		store := In(t.TempDir())
		if err := store.Save(File{}); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(store.Path(), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := store.Load()
		if err == nil || strings.Contains(err.Error(), secret) {
			t.Errorf("%s: %v; want an error that quotes no value", name, err)
		}
	}
}

func TestLoadRefusesALargeFile(t *testing.T) {
	store := In(t.TempDir())
	if err := store.Save(File{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.Path(), make([]byte, maxFileBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("%v; want too large", err)
	}
}

func TestLoadAndSaveFailOnAnUnusableDirectory(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "file")
	if err := os.WriteFile(blocked, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	store := In(blocked) // its epx directory would be below a file
	if err := store.Save(File{}); err == nil {
		t.Error("save below a file succeeded")
	}
	if _, err := store.Load(); err == nil {
		t.Error("load below a file succeeded")
	}
}

func TestCheckName(t *testing.T) {
	for name, want := range map[string]bool{
		"dev": true, "acme1-prod": true, "a.b_c": true, "9": true, strings.Repeat("a", 64): true,
		"": false, "-dev": false, ".": false, "a b": false, "a/b": false, strings.Repeat("a", 65): false,
		// Windows ignores case in the keychain's names, so a name has one case only.
		"Dev": false, "dEV": false,
	} {
		if err := CheckName(name); (err == nil) != want {
			t.Errorf("CheckName(%q) = %v; want valid %v", name, err, want)
		}
	}
}
