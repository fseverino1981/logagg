package cmd

import (
	"context"
	"fmt"
	"logagg/internal/aggregator"
	"logagg/internal/filter"
	"logagg/internal/reader"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var files []string
var filterParam string
var tail bool

var rootCmd = &cobra.Command{
	Use:   "logagg",
	Short: "Monitorador de logs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		channels := make([]<-chan string, 0, len(files))

		for _, f := range files {

			if err := reader.ValidateFile(f); err != nil {
				fmt.Println("Erro: ", err)
				continue
			}

			ch := reader.ReadLines(ctx, f, tail)
			channels = append(channels, ch)

		}

		lines := aggregator.Aggregate(ctx, channels...)
		result := filter.Filter(lines, filterParam)

		for l := range result {
			fmt.Println(l)
		}

	},
}

func init() {

	rootCmd.Flags().StringSliceVarP(&files, "files", "f", []string{}, "Arquivos para monitorar")
	rootCmd.Flags().StringVarP(&filterParam, "filter", "F", "", "Filtar o retorno do log por palavra")
	rootCmd.Flags().BoolVarP(&tail, "tail", "t", false, "Aguarda novas linhas no arquivo de log")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
