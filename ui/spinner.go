package ui

import (
	"io"
	"sync"
	"time"
)

var anim = [...]rune{
	'⠾', '⠷', '⠯', '⠟', '⠻', '⠽',
}

func Spinner(f io.Writer) func() {
	status := make(chan string, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	status <- " Loading"

	go func() {
		ticker := time.NewTicker(time.Second / 10)
		defer ticker.Stop()
		defer wg.Done()
		var str string

		idx := 0
		for {
			f.Write([]byte("\033[K"))
			f.Write([]byte("\033[1000D"))
			f.Write([]byte("\033[36m"))
			f.Write([]byte(string(anim[idx])))
			f.Write([]byte("\033[0m "))
			f.Write([]byte(str))

			select {
			case _str, ok := <-status:
				if !ok {
					f.Write([]byte("\033[K"))
					f.Write([]byte("\033[1000D"))
					f.Write([]byte("\033[?25h"))
					return
				}
				str = _str
			case <-ticker.C:
				idx = (idx + 1) % len(anim)
			}
		}
	}()

	return func() {
		close(status)
		wg.Wait()
	}
}
