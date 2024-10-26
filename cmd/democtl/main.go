package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wzshiming/democtl/cmd/democtl/mp4"
	"github.com/wzshiming/democtl/cmd/democtl/record"
	"github.com/wzshiming/democtl/cmd/democtl/svg"
)

func main() {
	cmd := &cobra.Command{
		Args: cobra.NoArgs,
		Use:  "democtl [command]",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		record.NewCommand(),
		svg.NewCommand(),
		mp4.NewCommand(),
	)

	err := cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
