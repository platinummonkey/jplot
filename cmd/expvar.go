package cmd

import (
	"log"
	"sync"
	"time"

	"github.com/rs/jplot/data"
	"github.com/rs/jplot/source"
	"github.com/rs/jplot/window"
	"github.com/spf13/cobra"
)

var url string

// expvarCmd represents the expvar command
var expvarCmd = &cobra.Command{
	Use:   "expvar",
	Short: "Graph using expvar-like json http server",
	Long: `Graph using expvar-like json http server

Example: (Using the example producer in doc/)

    jplot expvar --url http://:8080/debug/vars mem.heap+mem.sys+mem.stack counter:cpu.sTime+cpu.uTime threads
`,
	Run: func(cmd *cobra.Command, args []string) {
		runExpvar(args)
	},
}

func init() {
	rootCmd.AddCommand(expvarCmd)

	expvarCmd.Flags().StringVar(&url, "url", "", "URL to fetch every second")
	expvarCmd.MarkFlagRequired("url")
}

func runExpvar(args []string) {
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

	var s source.Getter = source.NewStdin()
	if url != "" {
		s = source.NewHTTP(url, time.Second)
	}
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
				v, err := jq.Query(f.Name)
				if err != nil {
					log.Fatalf("Cannot get %s: %v", f.Name, err)
				}
				n, ok := v.(float64)
				if !ok {
					log.Fatalf("Invalid type %s: %T", f.Name, v)
				}
				dp.Push(f.ID, n, f.Counter)
			}
		}
	}
}

