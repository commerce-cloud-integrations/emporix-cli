package config

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestFileKeepsItsRules(t *testing.T) {
	var f File
	checkDefaults(t, &f)
	checkUnknownProfiles(t, &f)
	if !reflect.DeepEqual(f.Profiles, map[string]Profile{"dev": {Tenant: "acme1", ClientID: "c1"}, "ci": {}}) {
		t.Errorf("profiles %+v", f.Profiles)
	}
}

// checkDefaults checks that the first profile put becomes the default, that a
// later one does not, that a replacement replaces, and that the rules of use
// and removal hold for the default.
func checkDefaults(t *testing.T, f *File) {
	t.Helper()
	if replaced := f.Put("dev", &Profile{Tenant: "acme1", ClientID: "c1"}); replaced || f.Default != "dev" {
		t.Fatalf("first put: replaced %v, default %q; want a new default", replaced, f.Default)
	}
	if replaced := f.Put("prod", &Profile{Tenant: "acme2", ClientID: "c2"}); replaced || f.Default != "dev" {
		t.Fatalf("second put: replaced %v, default %q; want dev to stay the default", replaced, f.Default)
	}
	if replaced := f.Put("prod", &Profile{Tenant: "acme2", ClientID: "c3"}); !replaced {
		t.Fatal("a put of an existing profile did not replace it")
	}
	if got, err := f.Lookup("prod"); err != nil || got.ClientID != "c3" {
		t.Fatalf("lookup %+v, %v; want the replaced profile", got, err)
	}
	if err := f.Use("prod"); err != nil || f.Default != "prod" {
		t.Fatalf("use: %v, default %q", err, f.Default)
	}
	if err := f.Remove("prod"); err != nil || f.Default != "" || len(f.Profiles) != 1 {
		t.Fatalf("remove of the default: %v, %+v; want no default and dev only", err, f)
	}
	f.Put("ci", &Profile{})
	if f.Default != "ci" {
		t.Errorf("a put without a default left default %q; want ci", f.Default)
	}
}

// checkUnknownProfiles checks that lookup, use and remove of an unknown profile
// fail with ErrNoProfile.
func checkUnknownProfiles(t *testing.T, f *File) {
	t.Helper()
	for name, call := range map[string]func() error{
		"lookup": func() error { _, err := f.Lookup("nope"); return err },
		"use":    func() error { return f.Use("nope") },
		"remove": func() error { return f.Remove("nope") },
	} {
		if err := call(); !errors.Is(err, ErrNoProfile) || err.Error() != `no profile "nope"` {
			t.Errorf("%s of an unknown profile: %v; want ErrNoProfile", name, err)
		}
	}
}

func TestSaveRefusesWhatLoadRefuses(t *testing.T) {
	for name, f := range map[string]File{
		"an invalid name": {Profiles: map[string]Profile{"Dev": {}}},
		"a stray default": {Default: "dev", Profiles: map[string]Profile{}},
	} {
		store := In(t.TempDir())
		if err := store.Save(f); err == nil || !strings.HasPrefix(err.Error(), "the configuration names ") {
			t.Errorf("%s: %v; want a refusal", name, err)
		}
	}
}
