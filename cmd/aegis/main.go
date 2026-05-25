// Package main provides the aegis root application.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aegis-run/aegis/cmd/aegis/datastore"
	"github.com/aegis-run/aegis/cmd/aegis/serve"
	"github.com/aegis-run/aegis/pkg/cli"
)

func main() {
	out := cli.Default()

	var cfgFile string

	rootCmd := &cobra.Command{
		Use: "aegis",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if cfgFile != "" {
				viper.SetConfigFile(cfgFile)
			} else {
				viper.AddConfigPath(".")
				viper.SetConfigName("aegis")
			}

			viper.SetEnvPrefix("AEGIS")
			viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
			viper.AutomaticEnv()

			if err := viper.ReadInConfig(); err != nil {
				if cfgFile != "" {
					return fmt.Errorf("failed to read specific config file: %w", err)
				}
			}

			out.Configure()

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile,
		"config",
		"",
		"config file (default is `./aegis.yaml`)",
	)

	rootCmd.PersistentFlags().AddFlagSet(out.Flags())

	help := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		out.PrintBanner()
		help(cmd, args)
	})

	rootCmd.AddCommand(serve.Command())
	rootCmd.AddCommand(datastore.Command())

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		out.Fatal("%v", err)
	}
}
