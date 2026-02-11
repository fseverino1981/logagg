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

func ReadLines(ctx context.Context, file string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			time.Sleep(time.Millisecond * 10)
			select {
			case out <- fmt.Sprintf("[%s] - %s", filepath.Base(file), scanner.Text()):
			case <-ctx.Done():
				fmt.Println("Sistema encerrado")
				return
			}
		}
	}()
	return out
}
