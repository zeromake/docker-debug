package command

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RequiresMinArgs returns an error if there is not at least min args
func RequiresMinArgs(min int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) >= min {
			return nil
		}
		return errors.Errorf(
			"%q requires at least %d %s.\nSee '%s --help'.\n\nUsage:  %s\n\n%s",
			cmd.CommandPath(),
			min,
			pluralize("argument", min),
			cmd.CommandPath(),
			cmd.UseLine(),
			cmd.Short,
		)
	}
}
func pluralize(word string, number int) string {
	if number == 1 {
		return word
	}
	return word + "s"
}
