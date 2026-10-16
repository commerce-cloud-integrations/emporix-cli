package config

import (
	"errors"

	"github.com/zalando/go-keyring"
)

// ErrNoSecret is what a Keychain returns for a profile whose client secret
// it does not hold.
var ErrNoSecret = errors.New("no client secret stored")

// Keychain holds the client secret of each profile, by profile name.
type Keychain interface {
	// Secret returns the client secret of profile, or ErrNoSecret.
	Secret(profile string) (string, error)
	// SetSecret stores secret as the client secret of profile, replacing any.
	SetSecret(profile, secret string) error
	// DeleteSecret deletes the client secret of profile, or returns ErrNoSecret.
	DeleteSecret(profile string) error
}

// KeychainService is the service under which the OS keychain holds the
// client secrets of profiles, one account per profile.
const KeychainService = "epx"

// SystemKeychain is the keychain of the operating system: the macOS
// Keychain, the Windows Credential Manager, or a Secret Service (such as
// GNOME Keyring or KWallet) on Linux and BSD. It needs no cgo.
type SystemKeychain struct{}

// Secret implements Keychain.
func (SystemKeychain) Secret(profile string) (string, error) {
	secret, err := keyring.Get(KeychainService, profile)
	return secret, noSecret(err)
}

// SetSecret implements Keychain.
func (SystemKeychain) SetSecret(profile, secret string) error {
	return keyring.Set(KeychainService, profile, secret)
}

// DeleteSecret implements Keychain.
func (SystemKeychain) DeleteSecret(profile string) error {
	return noSecret(keyring.Delete(KeychainService, profile))
}

// noSecret returns err, with the keyring's not found as ErrNoSecret.
func noSecret(err error) error {
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrNoSecret
	}
	return err
}
