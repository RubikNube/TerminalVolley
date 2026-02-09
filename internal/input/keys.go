// Package input handles keyboard input for the game.
package input

import (
	"bufio"
	"os"
)

type Key byte

const (
	KeyUnknown Key = 0
	KeyQuit    Key = 'q'
)

func StartKeyReader() <-chan Key {
	ch := make(chan Key, 64)

	go func() {
		r := bufio.NewReaderSize(os.Stdin, 64)
		for {
			b, err := r.ReadByte()
			if err != nil {
				close(ch)
				return
			}
			ch <- Key(b)
		}
	}()

	return ch
}
