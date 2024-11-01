package svg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wzshiming/democtl/pkg/color"
	"github.com/wzshiming/democtl/pkg/renderer"
	"github.com/wzshiming/democtl/pkg/renderer/svg"
)

func NewCommand() *cobra.Command {
	var (
		input   string
		output  string
		profile string
	)
	cmd := &cobra.Command{
		Use:   "svg",
		Short: "Convert terminal session to svg",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("no input file specified")
			}
			err := run(input, output, profile)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&input, "input", "i", input, "input filename")
	cmd.Flags().StringVarP(&output, "output", "o", output, "output filename")
	cmd.Flags().StringVarP(&profile, "profile", "p", profile, "profile")
	return cmd
}

func run(inputPath, outputPath, profile string) (err error) {
	c := color.DefaultColors()
	if profile != "" {
		c, err = color.NewColorsFromFile(profile)
		if err != nil {
			return err
		}
	}

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

	err = renderer.Render(context.Background(), svg.NewCanvas(outputFile, false, c.GetColorForHex), input)
	if err != nil {
		return err
	}
	return nil
}
