package main

import (
	"logagg/cmd"
)

func main() {
	//ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	//defer cancel()

	cmd.Execute()
}
