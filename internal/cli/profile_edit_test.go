package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestProfileUseSetsTheDefault(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), nil)
	code, stdout, stderr := runIn(t, setup.env(nil, "", false), "profile", "use", "prod")
	if code != ExitOK || stdout != "" || stderr != "prod is the default profile.\n" || setup.load(t).Default != "prod" {
		t.Fatalf("exit %d, stdout %q, stderr %q, default %q", code, stdout, stderr, setup.load(t).Default)
	}
	if code, _, stderr := runIn(t, setup.env(nil, "", false), "profile", "use", "nope"); code != ExitUsage ||
		!strings.Contains(stderr, `no profile "nope"; `+"`epx profile list` lists them") ||
		setup.load(t).Default != "prod" {
		t.Errorf("unknown: exit %d, stderr %q", code, stderr)
	}
}

func TestProfileRemoveDeletesTheProfileAndItsSecret(t *testing.T) {
	fake := newFakeEmporix(t)
	for _, tt := range []struct {
		name     string
		args     []string
		stdin    string
		terminal bool
		code     int
		message  string
		removed  bool
	}{
		{name: "--yes", args: []string{"--yes"}, code: ExitOK, message: "Removed profile dev and its client secret.\n",
			removed: true},
		{name: "confirmed", stdin: "y\n", terminal: true, code: ExitOK,
			message: "Remove profile dev and its client secret? [y/N] Removed profile dev", removed: true},
		{name: "declined", stdin: "n\n", terminal: true, code: ExitError, message: "not confirmed; nothing was removed"},
		{name: "no terminal", code: ExitUsage, message: "epx profile remove asks for confirmation on a terminal, " +
			"which standard input is not: give --yes"},
	} {
		setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret, "prod": "p"})
		code, stdout, stderr := runIn(t, setup.env(nil, tt.stdin, tt.terminal), append([]string{"profile", "remove",
			"dev"}, tt.args...)...)
		file := setup.load(t)
		_, kept := file.Profiles["dev"]
		_, secretKept := setup.keychain.secrets["dev"]
		if code != tt.code || stdout != "" || !strings.Contains(stderr, tt.message) || kept == tt.removed ||
			secretKept == tt.removed {
			t.Errorf("%s: exit %d, stderr %q, profile kept %v, secret kept %v", tt.name, code, stderr, kept, secretKept)
		}
		if tt.removed && (file.Default != "" || setup.keychain.secrets["prod"] != "p") {
			t.Errorf("%s: default %q, keychain %q; want no default and prod's secret", tt.name, file.Default,
				setup.keychain.secrets)
		}
	}
}

func TestProfileRemoveFailures(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret})
	// A profile without a stored secret is removed.
	if code, _, stderr := runIn(t, setup.env(nil, "", false), "profile", "remove", "prod", "--yes"); code != ExitOK {
		t.Fatalf("without a secret: exit %d, stderr %q", code, stderr)
	}
	for _, args := range [][]string{{"--yes"}, {}} {
		code, _, stderr := runIn(t, setup.env(nil, "y\n", true), append([]string{"profile", "remove", "prod"}, args...)...)
		if code != ExitUsage || !strings.Contains(stderr, `no profile "prod"`) {
			t.Errorf("unknown %q: exit %d, stderr %q", args, code, stderr)
		}
	}
	// A keychain that cannot delete the secret keeps no profile, and says where the secret is.
	setup.keychain.fail = map[string]error{"DeleteSecret": errors.New("locked")}
	code, _, stderr := runIn(t, setup.env(nil, "", false), "profile", "remove", "dev", "--yes")
	want := "removed profile dev, but the OS keychain may still hold its client secret (service epx, account dev): locked"
	if _, kept := setup.load(t).Profiles["dev"]; code != ExitError || !strings.Contains(stderr, want) || kept {
		t.Errorf("keychain failure: exit %d, stderr %q, kept %v; want %q", code, stderr, kept, want)
	}
}

func TestProfileRemoveKeepsTheSecretWhenTheProfileCannotBeSaved(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret})
	readOnly(t, setup)
	code, _, stderr := runIn(t, setup.env(nil, "", false), "profile", "remove", "dev", "--yes")
	_, kept := setup.load(t).Profiles["dev"]
	if code != ExitError || !kept || setup.keychain.secrets["dev"] != testSecret {
		t.Errorf("exit %d, stderr %q, profile kept %v, keychain %q; want both kept", code, stderr, kept,
			setup.keychain.secrets)
	}
}

// TestProfileChangesWaitForEachOther checks that a profile added while
// another add waits for its secret keeps both: an add reads the
// configuration file only once it has the secret, under the file's lock.
func TestProfileChangesWaitForEachOther(t *testing.T) {
	setup := newProfileSetup(t, nil, nil)
	asked, answer := make(chan struct{}), make(chan string)
	env := setup.env(nil, "", true)
	env.Profiles.ReadHidden = func(context.Context) (string, error) {
		close(asked)
		return <-answer, nil
	}
	var out, errOut bytes.Buffer
	env.Out, env.Err = &out, &errOut
	first := make(chan int, 1)
	go func() {
		first <- Run(context.Background(), []string{"profile", "add", "dev", "--tenant", "acme1", "--client-id", "c1"}, env)
	}()
	<-asked
	if code, _, stderr := runIn(t, setup.env(nil, "s2", false), "profile", "add", "ci", "--tenant", "acme1",
		"--client-id", "c2"); code != ExitOK {
		t.Fatalf("second add: exit %d, stderr %q", code, stderr)
	}
	answer <- "s1"
	if code := <-first; code != ExitOK {
		t.Fatalf("first add: exit %d, stderr %q", code, errOut.String())
	}
	file := setup.load(t)
	if len(file.Profiles) != 2 || file.Default != "ci" || setup.keychain.secrets["dev"] != "s1" {
		t.Errorf("profiles %+v, keychain %q; want both, ci the default", file, setup.keychain.secrets)
	}
	// A change waits for one that holds the lock, until interrupted.
	unlock, err := setup.store().Lock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	errOut.Reset()
	if code := Run(ctx, []string{"profile", "use", "dev"}, env); code != ExitError ||
		!strings.Contains(errOut.String(), "locking the configuration file: waiting for another epx") {
		t.Errorf("while locked: exit %d, stderr %q", code, errOut.String())
	}
}

// TestAProfileReadWaitsForAReplacement checks that a command that reads a
// profile waits while another epx holds the configuration lock, which a
// replacement holds while it changes the profile and its secret: the profile
// and the secret are then read as one state, never the old host with the new
// secret.
func TestAProfileReadWaitsForAReplacement(t *testing.T) {
	setup := newProfileSetup(t, nil, nil)
	if code, _, stderr := runIn(t, setup.env(nil, "s1", false), "profile", "add", "dev", "--tenant", "acme1",
		"--client-id", "c1"); code != ExitOK {
		t.Fatalf("profile add: exit %d, stderr %q", code, stderr)
	}
	unlock, err := setup.store().Lock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	var errOut bytes.Buffer
	env := setup.env(nil, "", false)
	env.Err = &errOut
	if code := Run(ctx, []string{"auth", "status", "--profile", "dev"}, env); code != ExitError ||
		!strings.Contains(errOut.String(), "locking the configuration file") {
		t.Errorf("while locked: exit %d, stderr %q; want the read to wait for the lock", code, errOut.String())
	}
}

// TestAnInterruptEndsTheWaitForTheKeychain checks that a keychain call that
// does not return, as one waiting for a prompt the user cannot reach, ends
// with an interrupt.
func TestAnInterruptEndsTheWaitForTheKeychain(t *testing.T) {
	fake := newFakeEmporix(t)
	setup := newProfileSetup(t, twoProfiles(fake), map[string]string{"dev": testSecret})
	setup.keychain.block, setup.keychain.waiting = make(chan struct{}), make(chan struct{})
	t.Cleanup(func() { close(setup.keychain.block) })
	for _, args := range [][]string{
		{"auth", "status"}, {"profile", "show", "dev"}, {"profile", "remove", "dev", "--yes"},
		{"profile", "add", "ci", "--tenant", "acme1", "--client-id", "c"},
	} {
		env := setup.env(nil, "s", false)
		var out, errOut bytes.Buffer
		env.Out, env.Err = &out, &errOut
		ctx, cancel := context.WithCancel(context.Background())
		go func() { // interrupts once the keychain waits, however slow the command is to reach it
			select {
			case <-setup.keychain.waiting:
				cancel()
			case <-ctx.Done():
			}
		}()
		code := Run(ctx, args, env)
		cancel()
		if code != ExitError || !strings.Contains(errOut.String(), "waiting for the OS keychain: context canceled") {
			t.Errorf("%q: exit %d, stderr %q", args, code, errOut.String())
		}
	}
	fake.assertNothingSent(t)
}
