package svg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wzshiming/democtl/pkg/renderer"
	"github.com/wzshiming/democtl/pkg/renderer/svg"
	"github.com/wzshiming/democtl/pkg/styles"
)

func NewCommand() *cobra.Command {
	var (
		input          string
		output         string
		profile        string
		iterationCount string = "infinite"
	)
	cmd := &cobra.Command{
		Use:   "svg",
		Short: "Convert terminal session to svg",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("no input file specified")
			}
			err := run(cmd.Context(), input, output, profile, iterationCount)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&input, "input", "i", input, "input filename")
	cmd.Flags().StringVarP(&output, "output", "o", output, "output filename")
	cmd.Flags().StringVarP(&profile, "profile", "p", profile, "profile")
	cmd.Flags().StringVar(&iterationCount, "count", iterationCount, "iteration count")
	return cmd
}

func run(ctx context.Context, inputPath, outputPath, profile string, iterationCount string) (err error) {
	c := styles.Default()
	if profile != "" {
		c, err = styles.NewStylesFromFile(profile)
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

	canvas := svg.NewCanvas(outputFile,
		svg.WithIterationCount(iterationCount),
		svg.WithGetColor(c.GetColorForHex),
		svg.WithWindows(!c.NoWindows),
	)

	err = renderer.Render(ctx, canvas, input)
	if err != nil {
		return err
	}
	return nil
}
