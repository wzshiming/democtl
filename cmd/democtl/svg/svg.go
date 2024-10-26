package svg

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/wzshiming/democtl/pkg/minify"
	"github.com/wzshiming/democtl/pkg/svg"
	"path/filepath"
)

func NewCommand() *cobra.Command {
	var (
		input  string
		output string
	)
	cmd := &cobra.Command{
		Use:   "svg",
		Short: "Convert terminal session to svg",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("no input file specified")
			}
			err := run(input, output)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&input, "input", "i", input, "input filename")
	cmd.Flags().StringVarP(&output, "output", "o", output, "output filename")
	return cmd
}

func run(inputPath, outputPath string) error {
	input, err := os.OpenFile(inputPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer input.Close()

	if outputPath == "" {
		inputExt := filepath.Ext(inputPath)
		outputPath = inputPath[:len(inputPath)-len(inputExt)] + ".svg"
	}

	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	var output io.Writer = outputFile
	mout := minify.SVGWithWriter(output)
	defer mout.Close()
	output = mout
	c := svg.NewCanvas()
	err = c.Run(input, output, false)
	if err != nil {
		return err
	}
	return nil
}
