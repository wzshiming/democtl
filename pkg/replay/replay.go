package replay

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/wzshiming/democtl/pkg/cast"
)

func Replay(ctx context.Context, input io.Reader) error {
	decoder := cast.NewDecoder(input)
	_, err := decoder.DecodeHeader()
	if err != nil {
		return err
	}

	lastTime := 0.
	for ctx.Err() == nil {
		event, err := decoder.DecodeEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		time.Sleep(time.Duration((event.Time - lastTime) * float64(time.Second)))
		lastTime = event.Time

		_, err = os.Stdout.WriteString(event.Data)
		if err != nil {
			return err
		}
	}

	return nil
}
