package serve

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aegis-run/aegis/internal/aegis"
	"github.com/aegis-run/aegis/internal/aegis/config"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "serve the Aegis server",
		RunE:  serve,
		Args:  cobra.NoArgs,
	}

	cmd.Flags().AddFlagSet(config.Flags())
	cmd.SilenceUsage = true

	return cmd
}

func serve(cmd *cobra.Command, _ []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	cfg, err := config.Load(viper.GetViper())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return aegis.Run(cmd.Context(), &cfg)
}
