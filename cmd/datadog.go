package cmd

import (
	"log"
	"sync"
	"time"
	"os"

	"github.com/rs/jplot/data"
	"github.com/rs/jplot/window"
	"github.com/rs/jplot/source"
	"github.com/spf13/cobra"
)

var datadogApiKey string

// datadogCmd represents the datadog command
var datadogCmd = &cobra.Command{
	Use:   "datadog",
	Short: "Graph using datadog",
	Long: `Graph using datadog metrics

Example:
    jplot datadog --key 123412341234123412341234 mem.heap+mem.sys+mem.stack counter:cpu.sTime+cpu.uTime threads
`,
	Run: func(cmd *cobra.Command, args []string) {
		runDatadog(args)
	},
}

func init() {
	rootCmd.AddCommand(datadogCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// datadogCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// datadogCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	datadogCmd.Flags().StringVar(&datadogApiKey, "key", "", "Datadog API Key (obtain through the UI)")
	datadogCmd.MarkFlagRequired("key")
}

func runDatadog(args []string) {
	specs := parseSpec(args)

	dp := &data.Points{Size: NumberPoints}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	exit := make(chan struct{})
	defer close(exit)
	go func() {
		defer wg.Done()
		window.Clear()
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			width, height, err := window.Size()
			if err != nil {
				log.Fatal("Cannot get window size")
			}
			select {
			case <-t.C:
				window.Render(specs, dp, width, height-25)
			case <-exit:
				window.Render(specs, dp, width, height-25)
				return
			}
		}
	}()

	if datadogApiKey == "" {
		log.Fatal("Invalid api key specified")
		os.Exit(1)
	}
	var s source.Getter = source.NewDatadog(datadogApiKey, specs, time.Second)
	defer s.Close()
	for {
		jq, err := s.Get()
		if err != nil {
			log.Fatalf("Input error: %v", err)
		}
		if jq == nil {
			break
		}
		for _, gs := range specs {
			for _, f := range gs.Fields {
				dataPoints, err := jq.Query(f.Name)
				if err != nil {
					log.Fatalf("Cannot get %s: %v", f.Name, err)
				}
				d, ok := dataPoints.([][]float64)
				if !ok {
					log.Fatalf("Invalid type %s: %T", f.Name, dataPoints)
				}
				// we only push the last value for now
				if len(d) > 0 {
					dp.Push(f.ID, d[len(d)-1][1], f.Counter)
				}
			}
		}
	}
}
