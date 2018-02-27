package source

import (
	"fmt"
	"time"

	"github.com/rs/jplot/data"
	"gopkg.in/zorkian/go-datadog-api.v2"
)

type DatadogSource struct {
	apiKey         string
	applicationKey string
	baseUrl        string
	client         *datadog.Client

	specs []data.GraphSpec
	c     chan Result
	done  chan struct{}

	// state
	lastQueryTime int64
}

func NewDatadog(apiKey, applicationKey, baseUrl string, specs []data.GraphSpec, interval time.Duration) *DatadogSource {
	client := datadog.NewClient(apiKey, applicationKey)
	if baseUrl != "" {
		client.SetBaseUrl(baseUrl)
	}
	s := &DatadogSource{
		apiKey:         apiKey,
		applicationKey: applicationKey,
		baseUrl:        baseUrl,
		client:         client,
		specs:          specs,
		c:              make(chan Result),
		done:           make(chan struct{}),
		lastQueryTime:  time.Now().Add(-time.Minute * 2).Unix(),
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
	maxUpdateTimestamp := s.lastQueryTime
	dataPoints := make(map[string]data.Points, 0)
	var err error
	for _, spec := range s.specs {
		for _, field := range spec.Fields {
			query := s.formatQuery(field)
			series, err := s.client.QueryMetrics(s.lastQueryTime, time.Now().Unix(), query)
			dataPoints[field.ID] = make(data.Points, 0)
			if err == nil && len(series) > 0 {
				endTs := int64(series[0].GetEnd() / 1000)
				if endTs > maxUpdateTimestamp {
					maxUpdateTimestamp = endTs
				}
				// assume the last data point is hte latest
				for _, ser := range series {
					for _, dp := range ser.Points {
						dataPoints[field.ID] = append(dataPoints[field.ID], data.Point{
							Timestamp: time.Unix(0, int64(dp[0])),
							Value:     dp[1],
						})
					}
				}
			} else if err != nil {
				break
			}
		}
	}

	if err != nil {
		s.c <- Result{Err: err}
		return
	}
	s.c <- Result{DataPoints: dataPoints, Err: err}
}

func (s *DatadogSource) formatQuery(field data.Field) string {
	querySuffix := ""
	if field.Counter {
		querySuffix = ".as_count()"
	}
	return fmt.Sprintf("avg:%s%s", field.Name, querySuffix)
}

func (s *DatadogSource) Close() error {
	close(s.done)
	return nil
}

func (s *DatadogSource) Get() (*Result, error) {
	res := <-s.c
	if res.Err != nil {
		return nil, res.Err
	}
	return &res, nil
}
