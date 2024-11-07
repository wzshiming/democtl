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
	"strconv"
	"time"

	"github.com/creack/pty"
	"github.com/google/shlex"
	"github.com/wzshiming/democtl/pkg/cast"
	"github.com/wzshiming/getch"
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

	ptmx           *os.File
	bufferedReader *bufferedReader

	typingInterval time.Duration
}

func NewPlayer(shell string, rows, cols uint16) *Player {
	return &Player{
		buffer:         make([]byte, 1024),
		shell:          shell,
		debug:          os.Stdout,
		rows:           rows,
		cols:           cols,
		typingInterval: time.Second / 10,
	}
}

func (p *Player) readOutput(timeout time.Duration) error {
	n, retErr := p.readWithTimeout(p.buffer, timeout)
	if retErr != nil && retErr != io.EOF {
		return retErr
	}

	now := time.Now()

	err := p.record(p.buffer[:n], now.UnixMicro())
	if err != nil {
		return err
	}

	return retErr
}

func (p *Player) record(b []byte, baseTime int64) error {
	_, err := p.debug.Write(b)
	if err != nil {
		return err
	}

	if p.baseTime == 0 {
		p.baseTime = baseTime
	}

	p.pushHistory(b)

	event := cast.Event{
		Time: float64(baseTime-p.baseTime) / float64(time.Millisecond),
		Data: string(b),
	}

	err = p.encoder.EncodeEvent(event)
	if err != nil {
		return err
	}
	return nil
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
	prompt2, err := p.getPrompt(target, time.Second/2)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(prompt1, prompt2) {
		return prompt1, nil
	}
	prompt3, err := p.getPrompt(target, time.Second/2)
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
		err := p.readOutput(time.Second)
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

	first := true
	reader := bufio.NewReader(in)
	for {
		if first {
			err = p.record(prompt, time.Now().UnixMicro())
			if err != nil {
				return err
			}
			first = false
		} else {
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

			c, err := p.builtinCommand(line)
			if err != nil {
				return err
			}
			if c {
				continue
			}

			haveNext = bytes.HasSuffix(line, []byte{'\\'})
			err = p.command(line)
			if err != nil {
				return err
			}
		}
	}
}

func (p *Player) builtinCommand(line []byte) (bool, error) {
	if !bytes.HasPrefix(line, []byte{'@'}) {
		return false, nil
	}
	args, err := shlex.Split(string(line[1:]))
	if err != nil {
		return false, err
	}
	switch args[0] {
	case "pause":
		_, _, _ = getch.Getch()
	case "sleep":
		if len(args) != 2 {
			return false, fmt.Errorf("sleep expects 2 arguments, got %d", len(args))
		}
		sleepDuration, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return false, err
		}
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	case "typing-interval":
		if len(args) != 2 {
			return false, fmt.Errorf("typing-interval expects 2 arguments, got %d", len(args))
		}
		interval, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return true, err
		}
		p.typingInterval = time.Duration(interval) * time.Second
	default:
		return false, fmt.Errorf("unknown command: %s", args[0])
	}
	return true, nil
}

func (p *Player) command(line []byte) error {
	time.Sleep(p.typingInterval)
	for i := range line {
		_, err := p.ptmx.Write(line[i : i+1])
		if err != nil {
			return err
		}
		err = p.readOutput(time.Second)
		if err != nil {
			return err
		}
		time.Sleep(p.typingInterval)
	}
	_, err := p.ptmx.Write([]byte{'\n'})
	if err != nil {
		return err
	}
	return nil
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
	p.bufferedReader = newBufferedReader(ptmx)
	go p.bufferedReader.Run()
	return p.run(in)
}

func (p *Player) readWithTimeout(buffer []byte, timeout time.Duration) (int, error) {
	return p.bufferedReader.ReadWithTimeout(buffer, timeout)
}
