package cmd

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/rs/jplot/data"
	"github.com/rs/jplot/source"
	"github.com/rs/jplot/window"
	"github.com/spf13/cobra"
)

var datadogApiKey string
var datadogApplicationKey string
var datadogBaseUrl string

// datadogCmd represents the datadog command
var datadogCmd = &cobra.Command{
	Use:   "datadog",
	Short: "Graph using datadog",
	Long: `Graph using datadog metrics

Must provide your Application Key

Example:
    jplot datadog --apiKey 123412341234123412341234 --appKey 123412341234123412341234 mem.heap+mem.sys+mem.stack counter:cpu.sTime+cpu.uTime threads
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
	datadogCmd.Flags().StringVar(&datadogApiKey, "apiKey", "", "Datadog API Key (obtain through the UI)")
	datadogCmd.MarkFlagRequired("apiKey")
	datadogCmd.Flags().StringVar(&datadogApplicationKey, "appKey", "", "Datadog Application Key (obtain through the UI)")
	datadogCmd.MarkFlagRequired("appKey")
	datadogCmd.Flags().StringVar(&datadogBaseUrl, "baseUrl", "", "Datadog Base URL (Defaults to https://app.datadoghq.com)")
}

func runDatadog(args []string) {
	specs := parseSpec(args)

	ds := &data.DataSet{
		Size:              NumberPoints,
		ExpectedFrequency: time.Second,
	}
	ready := NewAtomicReady(false)
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
				if ready.Ready() {
					window.Render(specs, ds, width, height-25)
				}
			case <-exit:
				if ready.Ready() {
					window.Render(specs, ds, width, height-25)
				}
				return
			}
		}
	}()

	if datadogApiKey == "" {
		log.Fatal("Invalid api key specified")
		os.Exit(1)
	}
	s := source.NewDatadog(datadogApiKey, datadogApplicationKey, datadogBaseUrl, specs, time.Second*10)
	defer s.Close()

	readyMap := make(map[string]bool, 0)
	for _, gs := range specs {
		for _, f := range gs.Fields {
			readyMap[f.ID] = false
		}
	}

	for {
		result, err := s.Get()
		if err != nil {
			log.Fatalf("Input error: %v", err)
		}
		if result == nil {
			break
		}
		for _, gs := range specs {
			for _, f := range gs.Fields {

				dataPoints, ok := result.DataPoints[f.ID]
				if !ok {
					log.Fatalf("Cannot get %s: %v", f.Name, err)
				}
				if len(dataPoints) > 0 {
					ds.PushPoints(f.ID, dataPoints, f.Counter)
					readyMap[f.ID] = true
				}
			}
		}
		if !ready.Ready() {
			// check if ready
			for _, v := range readyMap {
				if !v {
					continue
				}
			}
			ready.MarkReady()
		}
	}
}
