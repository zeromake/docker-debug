package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init()  {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "docker and client info",
		Args:  RequiresMinArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Test Info")
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}