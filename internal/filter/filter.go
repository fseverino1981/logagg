package filter

import "strings"

func Filter(ch <-chan string, filter string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		for f := range ch {
			if strings.Contains(f, filter) {
				out <- f
			}
		}
	}()

	return out

}
