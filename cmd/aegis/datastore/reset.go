package datastore

import (
	"github.com/spf13/cobra"

	"github.com/aegis-run/aegis/pkg/db/migrate"
)

func resetCommand() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:          "reset",
		Short:        "Reset database migration state",
		RunE:         runReset(&yes),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}

	cmd.Flags().AddFlagSet(migrate.Flags())
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "confirm destructive reset operation")

	return cmd
}

func runReset(yes *bool) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if err := requireYes(*yes, "reset"); err != nil {
			return err
		}

		return runWithConfig(cmd, "reset", migrate.Reset)
	}
}
