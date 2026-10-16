// Command epx is a command-line client for the Emporix Commerce APIs.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/cli"
	"github.com/commerce-cloud-integrations/emporix-cli/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := cli.Run(ctx, os.Args[1:], cli.Env{
		In: os.Stdin, Out: os.Stdout, Err: os.Stderr, Getenv: os.Getenv, Terminal: cli.IsTerminal(os.Stdin),
		Profiles: &cli.ProfileEnv{
			ConfigDir: os.UserConfigDir, Keychain: config.SystemKeychain{}, ReadHidden: cli.ReadHidden(os.Stdin),
		},
		OutTerminal: cli.IsTerminal(os.Stdout),
	})
	stop()
	os.Exit(code)
}
