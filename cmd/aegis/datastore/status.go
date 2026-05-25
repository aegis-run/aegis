package datastore

import (
	"github.com/spf13/cobra"

	"github.com/aegis-run/aegis/pkg/db/migrate"
)

func statusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        "Show database migration status",
		RunE:         runStatus,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}

	cmd.Flags().AddFlagSet(migrate.Flags())

	return cmd
}

func runStatus(cmd *cobra.Command, _ []string) error {
	return runWithConfig(cmd, "status", migrate.Status)
}
