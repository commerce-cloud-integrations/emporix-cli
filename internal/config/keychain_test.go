package config

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestSystemKeychainStoresSecretsPerProfile runs SystemKeychain against the
// keyring's in-memory mock, never a real keychain: each profile has its own
// secret, and a missing one is ErrNoSecret.
func TestSystemKeychainStoresSecretsPerProfile(t *testing.T) {
	keyring.MockInit()
	var k Keychain = SystemKeychain{}
	if _, err := k.Secret("dev"); !errors.Is(err, ErrNoSecret) {
		t.Fatalf("secret before set: %v; want ErrNoSecret", err)
	}
	if err := k.SetSecret("dev", "s1"); err != nil {
		t.Fatal(err)
	}
	if err := k.SetSecret("prod", "s2"); err != nil {
		t.Fatal(err)
	}
	if got, err := k.Secret("dev"); err != nil || got != "s1" {
		t.Fatalf("secret %q, %v; want s1", got, err)
	}
	if err := k.DeleteSecret("dev"); err != nil {
		t.Fatal(err)
	}
	if err := k.DeleteSecret("dev"); !errors.Is(err, ErrNoSecret) {
		t.Fatalf("second delete: %v; want ErrNoSecret", err)
	}
	if got, err := k.Secret("prod"); err != nil || got != "s2" {
		t.Fatalf("other profile's secret %q, %v; want s2", got, err)
	}
}

func TestSystemKeychainPassesOtherErrorsOn(t *testing.T) {
	failure := errors.New("keychain locked")
	keyring.MockInitWithError(failure)
	if _, err := (SystemKeychain{}).Secret("dev"); !errors.Is(err, failure) {
		t.Fatalf("error %v; want the keychain's", err)
	}
}
