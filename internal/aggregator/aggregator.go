package aggregator

import (
	"context"
	"sync"
)

func Aggregate(ctx context.Context, channels ...<-chan string) chan string {
	out := make(chan string)
	var wg sync.WaitGroup

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
