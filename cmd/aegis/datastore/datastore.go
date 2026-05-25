package datastore

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "datastore",
		Short:        "Manage the datastore",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}

	cmd.AddCommand(migrateCommand())
	cmd.AddCommand(statusCommand())
	cmd.AddCommand(resetCommand())

	return cmd
}
