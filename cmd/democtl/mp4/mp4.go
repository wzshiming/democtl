package mp4

import (
	"fmt"
	"os"
	"path/filepath"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/wzshiming/democtl/pkg/video"
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

	stat, err := os.Stat(outputPath)
	if err == nil {
		if stat.IsDir() {
			return fmt.Errorf("output directory already exists")
		} else {
			os.Remove(outputPath)
		}
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Printf(`# Next step: run the following command to generate the video
###############################
ffmpeg \
  -f concat \
  -safe 0 \
  -i %q \
  -vsync vfr \
  -pix_fmt yuv420p \
  %q
rm -rf %q
###############################
`, filepath.Join(rawDir, "frames.txt"), outputPath, rawDir)
		return nil
	}

	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", filepath.Join(rawDir, "frames.txt"),
		"-vsync", "vfr",
		"-pix_fmt", "yuv420p",
		outputPath,
	}

	err = exec.Command(ffmpegPath, args...).Run()
	if err != nil {
		return err
	}

	_, err = os.Stat(outputPath)
	if err != nil {
		return err
	}

	err = os.RemoveAll(rawDir)
	if err != nil {
		return err
	}

	return nil
}
