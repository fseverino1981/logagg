package cmd

import (
	"context"
	"fmt"
	"logagg/internal/aggregator"
	"logagg/internal/reader"
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/cobra"
)

var files []string

var rootCmd = &cobra.Command{
	Use:   "logagg",
	Short: "Monitorador de logs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		var wg sync.WaitGroup
		defer cancel()
		channels := make([]<-chan string, 0, len(files))
		fmt.Println(files)
		for _, f := range files {

			if err := reader.ValidateFile(f); err != nil {
				fmt.Println("Erro: ", err)
				continue
			}
			ch := reader.ReadLines(ctx, f)
			channels = append(channels, ch)

		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			for lines := range aggregator.Aggregate(ctx, channels...) {
				fmt.Println(lines)
			}
		}()
		wg.Wait()
		fmt.Println("Logs processados")
	},
}

func init() {

	rootCmd.Flags().StringSliceVarP(&files, "files", "f", []string{}, "Arquivos para monitorar")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
