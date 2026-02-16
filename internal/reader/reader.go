package reader

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func ReadLines(ctx context.Context, file string, tail bool) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)

		for {
			for scanner.Scan() {
				select {
				case out <- fmt.Sprintf("[%s] - %s", filepath.Base(file), scanner.Text()):
				case <-ctx.Done():
					return
				}
			}

			if !tail {
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
				scanner = bufio.NewScanner(f)
			}
		}
	}()

	return out

}
