package cmd

import (
	"context"
	"fmt"
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
		var wg sync.WaitGroup
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		channels := make([]chan string, len(files))
		for _, f := range files {
			if err := reader.ValidateFile(f); err != nil {
				fmt.Println("Erro: ", err)
				continue
			}
			wg.Add(1)
			go func(filename string) {
				defer wg.Done()
				for lines := range reader.ReadLines(ctx, filename) {
					fmt.Println(lines)
				}

			}(f)
		}
		wg.Wait()
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
