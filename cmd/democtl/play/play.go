package play

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wzshiming/democtl/pkg/replay"
)

func NewCommand() *cobra.Command {
	var (
		input string
	)

	cmd := &cobra.Command{
		Use:     "play",
		Aliases: []string{"play"},
		Short:   "Play terminal session",
		Args:    cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("no input file specified")
			}
			err := run(input)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&input, "input", "i", input, "input filename")
	return cmd
}

func run(inputPath string) error {
	input, err := os.OpenFile(inputPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer input.Close()
	err = replay.Replay(context.Background(), input)
	if err != nil {
		return err
	}
	return nil
}
