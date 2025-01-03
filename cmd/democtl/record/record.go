package record

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wzshiming/democtl/pkg/player"
)

func NewCommand() *cobra.Command {
	var (
		rows   uint16 = 24
		cols   uint16 = 86
		input  string
		output string
		shell  = os.Getenv("SHELL")
	)
	if shell == "" {
		shell = "sh"
	}
	cmd := &cobra.Command{
		Use:     "record",
		Aliases: []string{"rec"},
		Short:   "Record terminal session",
		Args:    cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("no input file specified")
			}
			err := run(input, output, shell, rows, cols)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().Uint16VarP(&rows, "rows", "r", rows, "number of rows")
	cmd.Flags().Uint16VarP(&cols, "cols", "c", cols, "number of columns")
	cmd.Flags().StringVarP(&input, "input", "i", input, "input filename")
	cmd.Flags().StringVarP(&output, "output", "o", output, "output filename")
	cmd.Flags().StringVarP(&shell, "shell", "s", shell, "shell script")
	return cmd
}

func run(inputPath, outputPath, shell string, rows, cols uint16) error {
	input, err := os.OpenFile(inputPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer input.Close()

	if outputPath == "" {
		inputExt := filepath.Ext(inputPath)
		outputPath = inputPath[:len(inputPath)-len(inputExt)] + ".cast"
	}

	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	p := player.NewPlayer(shell, rows, cols)
	err = p.Run(context.Background(), input, outputFile, filepath.Dir(inputPath))
	if err != nil {
		return err
	}
	return nil
}
