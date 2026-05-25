package datastore

import (
	"github.com/spf13/cobra"

	"github.com/aegis-run/aegis/pkg/db/migrate"
)

func migrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "migrate",
		Short:        "Run database migrations",
		RunE:         runMigrate,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}

	cmd.Flags().AddFlagSet(migrate.Flags())

	return cmd
}

func runMigrate(cmd *cobra.Command, _ []string) error {
	return runWithConfig(cmd, "migrate", migrate.Migrate)
}
