package datastore

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aegis-run/aegis/pkg/cli"
	"github.com/aegis-run/aegis/pkg/db/migrate"
	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

func printStart(cmd *cobra.Command, action string, cfg migrate.Config) time.Time {
	out := cli.Default()
	out.SetWriter(cmd.OutOrStdout())
	// logger.SetHandler(slog.NewTextHandler(out.Writer(), nil))
	logger.SetHandler(cli.NewSlogHandler(out))

	out.Header("Datastore %s", action)
	out.Label("Engine", string(cfg.Engine))
	out.Label("Migrations", cfg.MigrationsTable)
	out.Label("URI", redactURI(cfg.URI))

	return time.Now()
}

func printDone(cmd *cobra.Command, action string, startedAt time.Time) {
	out := cli.Default()
	out.SetWriter(cmd.OutOrStdout())

	out.Success("%s completed", action)
	out.Subtle("Duration: %s", time.Since(startedAt).Round(time.Millisecond))
}

func redactURI(raw string) string {
	if raw == "" {
		return ""
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "[invalid-uri-redacted]"
	}

	if u.User != nil {
		username := u.User.Username()
		if username != "" {
			u.User = url.UserPassword(username, "***")
		} else {
			u.User = url.User("***")
		}
	}

	if u.RawQuery != "" {
		u.RawQuery = "redacted"
	}

	u.Path = "/..."

	return u.String()
}

func requireYes(yes bool, action string) error {
	if yes {
		return nil
	}

	return fmt.Errorf("%s is destructive; re-run with --yes to confirm", action)
}

func runWithConfig(
	cmd *cobra.Command,
	action string,
	run func(context.Context, *migrate.Config) error,
) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	cfg, err := migrate.Load(viper.GetViper())
	if err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}

	out := cli.Default()
	out.SetWriter(cmd.OutOrStdout())
	cliLogger := slog.New(cli.NewSlogHandler(out))
	migrate.SetLogger(cliLogger)

	startedAt := printStart(cmd, action, cfg)
	if err := run(cmd.Context(), &cfg); err != nil {
		return err
	}

	printDone(cmd, action, startedAt)
	return nil
}
