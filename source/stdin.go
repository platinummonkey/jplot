package source

import (
	"bufio"
	"os"

	"encoding/json"
	"github.com/rs/jplot/data"
	"strconv"
	"time"
)

type Stdin struct {
	scan *bufio.Scanner
}

func NewStdin() Stdin {
	return Stdin{bufio.NewScanner(os.Stdin)}
}

func (s Stdin) Get() (*Result, error) {
	if s.scan.Scan() {
		var m map[string]interface{}
		err := json.Unmarshal(s.scan.Bytes(), &m)
		if err != nil {
			return nil, err
		}
		now := time.Now()

		dataPoints := make(map[string]data.Points, 0)
		for k, v := range m {
			var value float64
			switch v.(type) {
			case int:
				value = float64(v.(int))
			case int64:
				value = float64(v.(int64))
			case float32:
				value = float64(v.(float32))
			case float64:
				value = v.(float64)
			case string:
				value, err = strconv.ParseFloat(v.(string), 64)
			default:
				continue
			}

			dataPoints[k] = append(dataPoints[k], data.Point{
				Timestamp: now,
				Value:     value,
			})
		}

		return &Result{DataPoints: dataPoints, Err: nil}, nil
	}
	return nil, s.scan.Err()
}

func (s Stdin) Close() error {
	return nil
}
