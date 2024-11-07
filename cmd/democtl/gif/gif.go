package gif

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wzshiming/democtl/pkg/renderer"
	"github.com/wzshiming/democtl/pkg/renderer/video"
	"github.com/wzshiming/democtl/pkg/styles"
)

func NewCommand() *cobra.Command {
	var (
		input   string
		output  string
		profile string
	)
	cmd := &cobra.Command{
		Use:   "gif",
		Short: "Convert terminal session to gif",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("no input file specified")
			}
			err := run(cmd.Context(), input, output, profile)
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

func run(ctx context.Context, inputPath, outputPath, profile string) (err error) {
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
		outputPath = inputPath[:len(inputPath)-len(inputExt)] + ".gif"
	}

	rawDir := outputPath + ".raw"

	err = os.MkdirAll(rawDir, 0755)
	if err != nil {
		return err
	}

	canvas := video.NewCanvas(rawDir,
		video.WithGetColor(c.GetColorForHex),
		video.WithWindows(!c.NoWindows),
	)

	err = renderer.Render(ctx, canvas, input)
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
		outputPath,
	}

	info, err := exec.Command(ffmpegPath, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w:%s", err, string(info))
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
