package cmd

import (
	"context"
	"fmt"
	"logagg/internal/aggregator"
	"logagg/internal/filter"
	"logagg/internal/reader"
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/cobra"
)

var files []string
var filterParam string

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

			lines := aggregator.Aggregate(ctx, channels...)
			result := filter.Filter(lines, filterParam)

			for l := range result {
				fmt.Println(l)
			}

		}()
		wg.Wait()
		fmt.Println("Logs processados")
	},
}

func init() {

	rootCmd.Flags().StringSliceVarP(&files, "files", "f", []string{}, "Arquivos para monitorar")
	rootCmd.Flags().StringVarP(&filterParam, "filter", "t", "", "Filtar o retorno do log por palavra")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
