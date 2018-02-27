package source

import (
	"fmt"
	"time"

	"github.com/elgs/gojq"
	"github.com/rs/jplot/data"
	"gopkg.in/zorkian/go-datadog-api.v2"
)

type DatadogSource struct {
	key string
	client *datadog.Client
	specs []data.GraphSpec
	c    chan res
	done chan struct{}

	// state
	lastQueryTime int64
}

func NewDatadog(apiKey string, specs []data.GraphSpec, interval time.Duration) *DatadogSource {
	client := datadog.NewClient(apiKey, "")
	s := &DatadogSource{
		key: apiKey,
		client: client,
		specs: specs,
		c: make(chan res),
		done: make(chan struct{}),
		lastQueryTime: time.Now().Unix(),
	}
	go s.run(interval)
	return s
}

func (s *DatadogSource) run(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s.fetch()
		case <-s.done:
			return
		}
	}
}

func (s *DatadogSource) fetch() {
	maxUpdateTimestamp := int64(s.lastQueryTime)
	dataPoints := make(map[string]datadog.DataPoint, 0)
	var err error
	for _, spec := range s.specs {
		for _, field := range spec.Fields {
			query := s.formatQuery(field)
			series, err := s.client.QueryMetrics(s.lastQueryTime, time.Now().Unix(), query)
			if err == nil && len(series) > 0 {
				endTs := int64(series[0].GetEnd()/1000)
				if endTs > maxUpdateTimestamp {
					maxUpdateTimestamp = endTs
				}
				// assume the last data point is hte latest
				dataPoints[field.ID] = series[0].Points[len(series[0].Points)-1]
			} else if err != nil {
				break
			}
		}
	}

	if err != nil {
		s.c <- res{err: err}
		return
	}
	jq := gojq.NewQuery(dataPoints)
	s.c <- res{jq: jq, err: err}
}

func (s *DatadogSource) formatQuery(field data.Field) string {
	querySuffix := ""
	if field.Counter {
		querySuffix = ".as_count()"
	}
	return fmt.Sprintf("%s%s", field.Name, querySuffix)
}

func (s *DatadogSource) Close() error {
	close(s.done)
	return nil
}

func (s *DatadogSource) Get() (*gojq.JQ, error) {
	res := <-s.c
	return res.jq, res.err
}
