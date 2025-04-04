package picserial

import (
	"bytes"
	"fmt"
	"sync"

	"go.bug.st/serial"

	"github.com/tbe-team/raybot/internal/config"
)

const readBufferSize = 64

type client struct {
	cfg config.Serial

	port serial.Port
	mode serial.Mode

	writeMu sync.Mutex
}

func newClient(cfg config.Serial) *client {
	mode := serial.Mode{
		BaudRate: cfg.BaudRate,
		DataBits: cfg.DataBits.Int(),
	}

	switch cfg.StopBits {
	case config.SerialStopBitsOne:
		mode.StopBits = serial.OneStopBit
	case config.SerialStopBitsOnePointFive:
		mode.StopBits = serial.OnePointFiveStopBits
	case config.SerialStopBitsTwo:
		mode.StopBits = serial.TwoStopBits
	}

	switch cfg.Parity {
	case config.SerialParityNone:
		mode.Parity = serial.NoParity
	case config.SerialParityOdd:
		mode.Parity = serial.OddParity
	case config.SerialParityEven:
		mode.Parity = serial.EvenParity
	}

	return &client{
		cfg:  cfg,
		mode: mode,
	}
}

func (c *client) Open() error {
	port, err := serial.Open(c.cfg.Port, &c.mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	if err := port.SetReadTimeout(c.cfg.ReadTimeout); err != nil {
		return fmt.Errorf("failed to set read timeout: %w", err)
	}

	c.port = port
	return nil
}

func (c *client) Close() error {
	return c.port.Close()
}

func (c *client) Write(data []byte) error {
	data = append([]byte(">"), data...)
	data = append(data, '\r', '\n')

	c.writeMu.Lock()
	_, err := c.port.Write(data)
	c.writeMu.Unlock()

	return err
}

// Read reads data from the serial port.
func (c *client) Read() ([]byte, error) {
	return c.read()
}

// read continuously reads from the port until a complete message is received.
// A complete message starts with '>' and ends with CR LF (\r\n).
// The message is returned without the prefix and suffix
func (c *client) read() ([]byte, error) {
	var res []byte
	messageStarted := false

	for {
		buf := make([]byte, readBufferSize)
		n, err := c.port.Read(buf)
		if err != nil {
			return nil, err
		}

		// Only append the bytes that were actually read
		chunk := buf[:n]

		// If we haven't found the start marker yet, look for it
		if !messageStarted {
			startIdx := bytes.IndexByte(chunk, '>')
			if startIdx >= 0 {
				// Found the start marker, only append from that point
				res = append(res, chunk[startIdx:]...)
				messageStarted = true
			}
		} else {
			// Already found start marker, append the whole chunk
			res = append(res, chunk...)
		}

		// Check if we have a complete message
		if messageStarted && len(res) > 0 && res[0] == '>' && bytes.HasSuffix(res, []byte("\r\n")) {
			// Remove the prefix and suffix
			res = res[1 : len(res)-2]
			// Remove null bytes
			res = bytes.Trim(res, "\x00")
			return res, nil
		}
	}
}
