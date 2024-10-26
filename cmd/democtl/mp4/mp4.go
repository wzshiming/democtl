package mp4

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wzshiming/democtl/pkg/video"
	"path/filepath"
)

func NewCommand() *cobra.Command {
	var (
		input  string
		output string
	)
	cmd := &cobra.Command{
		Use:   "mp4",
		Short: "Convert terminal session to mp4",
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
	cmd.Flags().StringVarP(&output, "output", "o", output, "output directory")
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
		outputPath = inputPath[:len(inputPath)-len(inputExt)] + ".mp4"
	}

	rawDir := outputPath + ".raw"

	err = os.MkdirAll(rawDir, 0755)
	if err != nil {
		return err
	}

	c, err := video.NewCanvas()
	if err != nil {
		return err
	}
	err = c.Run(input, rawDir, false)
	if err != nil {
		return err
	}

	fmt.Printf(`# Next step: run the following command to generate the video
###############################
ffmpeg \
  -f concat \
  -safe 0 \
  -i %s/frames.txt \
  -vsync vfr \
  -pix_fmt yuv420p \
  %s
rm -rf %s
###############################
`, rawDir, outputPath, rawDir)

	return nil
}
