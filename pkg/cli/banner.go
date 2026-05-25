//nolint:lll
package cli

import (
	"github.com/fatih/color"

	"github.com/aegis-run/aegis/internal"
)

func (o *Output) PrintBanner() {
	c := color.New(color.FgGreen).SprintFunc()
	title := color.New(color.FgCyan, color.Bold).SprintFunc()
	content := color.New(color.FgCyan).SprintFunc()
	subtle := color.New(color.FgHiWhite).SprintFunc()

	o.Raw("")
	o.Raw("  %s   %s (%s)", c("    ___       "), title("A E G I S"), internal.Version)
	o.Raw("  %s   %s", c("   /   | ___  "), content("Open-source Centralized Authorization System"))
	o.Raw("  %s     ", c("  / /| |/ _ \\"))
	o.Raw("  %s   %s", c(" / ___ /  __/ "), subtle("docs: ................... https://docs.aegis.dev"))
	o.Raw("  %s   %s", c("/_/  |_\\___/  "), subtle("github: ..... https://github.com/aegis-run/aegis"))
	o.Raw("  %s   %s", "              ", subtle("blog: ................... https://aegis.dev/blog"))
	o.Raw("")
}

// PrintBanner prints the application banner using the default output.
func PrintBanner() {
	defaultOutput.PrintBanner()
}
