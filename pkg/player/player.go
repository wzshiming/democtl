package player

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"github.com/wzshiming/democtl/pkg/cast"
	"golang.org/x/sys/unix"
)

type Player struct {
	buffer   []byte
	baseTime int64

	history []byte

	shell string
	debug io.Writer
	rows  uint16
	cols  uint16

	encoder *cast.Encoder

	ptmx *os.File
}

func NewPlayer(rows, cols uint16) *Player {
	return &Player{
		buffer: make([]byte, 1024),
		shell:  os.Getenv("SHELL"),
		debug:  os.Stdout,
		rows:   rows,
		cols:   cols,
	}
}

func (p *Player) readOutput() error {
	n, retErr := p.readWithTimeout(p.buffer, time.Second)

	now := time.Now()

	if retErr != nil && retErr != io.EOF {
		return retErr
	}

	_, err := p.debug.Write(p.buffer[:n])
	if err != nil {
		return err
	}

	if p.baseTime == 0 {
		p.baseTime = now.UnixMicro()
	}

	p.pushHistory(p.buffer[:n])

	event := cast.Event{
		Time: float64(now.UnixMicro()-p.baseTime) / float64(time.Millisecond),
		Data: string(p.buffer[:n]),
	}

	err = p.encoder.EncodeEvent(event)
	if err != nil {
		return err
	}
	return retErr
}

func (p *Player) clearHistory() {
	p.history = p.history[:0]
}

func (p *Player) pushHistory(data []byte) {
	p.history = append(p.history, data...)
}

func (p *Player) getHistory() []byte {
	return p.history
}

func (p *Player) getPrompt(target []byte, timeout time.Duration) ([]byte, error) {
	end := time.Now().Add(timeout)
	for {
		timeout := end.Sub(time.Now())
		if timeout < 0 {
			break
		}
		_, err := p.readWithTimeout(p.buffer, timeout)
		if err != nil {
			continue
		}
	}

	_, err := p.ptmx.Write(target)
	if err != nil {
		return nil, err
	}

	var off int
	end = time.Now().Add(timeout)
	for {
		timeout := end.Sub(time.Now())
		if timeout < 0 {
			if off != 0 {
				break
			}
			return nil, context.DeadlineExceeded
		}
		n, err := p.readWithTimeout(p.buffer[off:], timeout)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				break
			}
			return nil, err
		}
		off += n
	}

	buf := p.buffer[:off]
	l := bytes.LastIndexAny(buf, "\n\x0d")
	if l > 0 {
		buf = buf[l+1:]
	}
	prompt := make([]byte, len(buf))
	copy(prompt, buf)
	return prompt, nil
}

func (p *Player) mustGetPrompt(target []byte) ([]byte, error) {
	prompt1, err := p.getPrompt(target, time.Second/2)
	if err != nil {
		return nil, err
	}
	prompt2, err := p.getPrompt(target, time.Second/5)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(prompt1, prompt2) {
		return prompt1, nil
	}
	prompt3, err := p.getPrompt(target, time.Second/5)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(prompt2, prompt3) {
		return prompt2, nil
	}
	if bytes.Equal(prompt1, prompt3) {
		return prompt1, nil
	}
	return nil, errors.New(fmt.Sprintf("can't get prompt %s, %s, %s", prompt1, prompt2, prompt1))
}

func (p *Player) waitFinish() error {
	for {
		err := p.readOutput()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				break
			}
			return err
		}
	}
	return nil
}

func (p *Player) run(in io.Reader) error {
	prompt, err := p.mustGetPrompt([]byte{'\n'})
	if err != nil {
		return err
	}

	_, err = p.ptmx.Write([]byte{'\n'})
	if err != nil {
		return err
	}

	reader := bufio.NewReader(in)
	for {
		err = p.waitFinish()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if !bytes.HasSuffix(p.getHistory(), prompt) {
			time.Sleep(time.Second / 10)
			continue
		}
		p.clearHistory()

		haveNext := true
		for haveNext {
			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}

			haveNext = bytes.HasSuffix(line, []byte{'\\'})

			time.Sleep(time.Second / 10)
			for i := range line {
				_, err = p.ptmx.Write(line[i : i+1])
				if err != nil {
					return err
				}
				err = p.readOutput()
				if err != nil {
					return err
				}
				time.Sleep(time.Second / 10)
			}

			_, err = p.ptmx.Write([]byte{'\n'})
			if err != nil {
				return err
			}
		}
	}
}

func (p *Player) Run(ctx context.Context, in io.Reader, out io.Writer, dir string) error {
	p.encoder = cast.NewEncoder(out)
	err := p.encoder.EncodeHeader(cast.Header{
		Width:  int(p.cols),
		Height: int(p.rows),
	})
	if err != nil {
		return err
	}

	c := exec.CommandContext(ctx, p.shell)
	c.Dir = dir
	ptmx, err := pty.StartWithSize(c, &pty.Winsize{
		Rows: p.rows,
		Cols: p.cols,
	})
	if err != nil {
		return err
	}
	defer ptmx.Close()

	p.ptmx = ptmx

	return p.run(in)
}

func (p *Player) readWithTimeout(buffer []byte, timeout time.Duration) (int, error) {
	fd := int(p.ptmx.Fd())
	var readfds unix.FdSet
	readfds.Set(fd)

	tv := unix.NsecToTimeval(timeout.Nanoseconds())

	n, err := unix.Select(fd+1, &readfds, nil, nil, &tv)
	if err != nil {
		if errors.Is(err, unix.EINTR) {
			n, err = unix.Select(fd+1, &readfds, nil, nil, &tv)
			if err != nil {
				return n, err
			}
		}
	}

	if n == 0 {
		return 0, context.DeadlineExceeded
	}

	return unix.Read(fd, buffer)
}
