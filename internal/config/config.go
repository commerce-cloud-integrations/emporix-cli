// Package config reads and writes the configuration file of epx: named
// profiles, each a tenant, API URL, client ID and default site and language,
// and the profile used by default. A profile never holds a client secret:
// the OS keychain does (Keychain).
package config

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

// Profile is the settings that a named profile gives epx.
type Profile struct {
	Tenant   string `json:"tenant"`
	APIURL   string `json:"apiUrl,omitempty"`
	ClientID string `json:"clientId"`
	Site     string `json:"site,omitempty"`
	Language string `json:"language,omitempty"`
}

// File is the content of the configuration file: the profiles by name and
// the name of the default profile, if any.
type File struct {
	Default  string             `json:"default,omitempty"`
	Profiles map[string]Profile `json:"profiles"`
}

// namePattern matches a profile name: a lowercase letter or digit, then up
// to 63 lowercase letters, digits, dots, underscores or hyphens. A name is
// also the account of its client secret in the OS keychain, which on Windows
// ignores case: two names that differ in case only would share one secret.
var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

// errName is what CheckName returns for a name that cannot name a profile.
var errName = errors.New("a profile name is a lowercase letter or digit, then up to 63 lowercase letters, " +
	"digits, '.', '_' or '-'")

// CheckName returns an error unless name can name a profile.
func CheckName(name string) error {
	if !namePattern.MatchString(name) {
		return errName
	}
	return nil
}

// ErrNoProfile is what a File returns for a name that names none of its
// profiles.
var ErrNoProfile = errors.New("no profile")

// Lookup returns the profile name, or ErrNoProfile.
func (f File) Lookup(name string) (Profile, error) {
	profile, ok := f.Profiles[name]
	if !ok {
		return Profile{}, fmt.Errorf("%w %q", ErrNoProfile, name)
	}
	return profile, nil
}

// Put adds the profile name, or replaces the profile of that name, and
// reports which. Where no profile is the default, it becomes the default.
func (f *File) Put(name string, profile *Profile) (replaced bool) {
	if f.Profiles == nil {
		f.Profiles = map[string]Profile{}
	}
	_, replaced = f.Profiles[name]
	f.Profiles[name] = *profile
	f.Default = cmp.Or(f.Default, name)
	return replaced
}

// Use makes the profile name the default, or returns ErrNoProfile.
func (f *File) Use(name string) error {
	if _, err := f.Lookup(name); err != nil {
		return err
	}
	f.Default = name
	return nil
}

// Remove removes the profile name, or returns ErrNoProfile. Removing the
// default profile leaves none.
func (f *File) Remove(name string) error {
	if _, err := f.Lookup(name); err != nil {
		return err
	}
	delete(f.Profiles, name)
	if f.Default == name {
		f.Default = ""
	}
	return nil
}

// check returns an error unless f is a configuration: every profile name is
// one that CheckName accepts, and the default, if any, names a profile.
// Errors quote no value.
func (f File) check() error {
	for name := range f.Profiles {
		if CheckName(name) != nil {
			return errors.New("names a profile with characters other than lowercase letters, digits, '.', '_' and '-'")
		}
	}
	if _, err := f.Lookup(f.Default); f.Default != "" && err != nil {
		return errors.New("names a default profile that it does not hold")
	}
	return nil
}

// maxFileBytes bounds the configuration file, far above any set of profiles.
const maxFileBytes = 1 << 20

// Store is the configuration file of epx in a user's configuration directory.
type Store struct{ dir string }

// In returns the store in configDir, the user's configuration directory
// (os.UserConfigDir): the file epx/config.json there.
func In(configDir string) Store { return Store{dir: filepath.Join(configDir, "epx")} }

// Path returns the path of the configuration file.
func (s Store) Path() string { return filepath.Join(s.dir, "config.json") }

// Load reads the configuration file; one that does not exist holds no
// profiles. A file that other users can read or change is refused, and so is
// one that is not a configuration: invalid JSON, an unknown member (such as a
// client secret), an invalid profile name or a default that names no
// profile. Errors never quote the file's values.
func (s Store) Load() (File, error) {
	file, err := os.Open(s.Path())
	if errors.Is(err, fs.ErrNotExist) {
		return File{Profiles: map[string]Profile{}}, nil
	}
	if err != nil {
		return File{}, err
	}
	defer func() { _ = file.Close() }() // only read
	info, err := file.Stat()
	if err != nil {
		return File{}, err
	}
	if shared(info.Mode()) {
		return File{}, fmt.Errorf("%s can be read or changed by other users: make it readable by you only (chmod 600)",
			s.Path())
	}
	data, err := io.ReadAll(io.LimitReader(file, maxFileBytes+1))
	if err != nil {
		return File{}, err
	}
	if len(data) > maxFileBytes {
		return File{}, fmt.Errorf("%s is larger than %d bytes", s.Path(), maxFileBytes)
	}
	return parse(data, s.Path())
}

// shared reports whether a file of mode grants other users any access. On
// Windows, whose access control lists Go does not report in a mode, the
// user's configuration directory is the user's own.
func shared(mode fs.FileMode) bool {
	return runtime.GOOS != "windows" && mode.Perm()&0o077 != 0
}

// parse returns the configuration that data holds.
func parse(data []byte, path string) (File, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var f File
	// The decoder's messages may quote a value, so they are not passed on.
	if err := decoder.Decode(&f); err != nil {
		return File{}, fmt.Errorf("%s is not an epx configuration: a JSON object of profiles "+
			"(tenant, apiUrl, clientId, site, language) and a default", path)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return File{}, fmt.Errorf("%s holds more than one JSON value", path)
	}
	if f.Profiles == nil {
		f.Profiles = map[string]Profile{}
	}
	if err := f.check(); err != nil {
		return File{}, fmt.Errorf("%s %w", path, err)
	}
	return f, nil
}

// Save replaces the configuration file with f, readable and writable by the
// user only, in a directory only the user can enter. The file is replaced
// whole, so a failure leaves the previous one. A File that Load would refuse
// is not saved.
func (s Store) Save(f File) error {
	if err := f.check(); err != nil {
		return fmt.Errorf("the configuration %w", err)
	}
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(s.dir, "config-*.json") // mode 0600
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(temp.Name()) }() // gone once renamed
	if _, err := temp.Write(append(data, '\n')); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), s.Path())
}
