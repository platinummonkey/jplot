package source

import (
	"encoding/json"
	"fmt"
	"github.com/rs/jplot/data"
	"io"
	"strconv"
	"time"
)

type Result struct {
	DataPoints map[string]data.Points
	Err        error
}

type Getter interface {
	io.Closer
	Get() (*Result, error)
}

func jsonSubMapAppendData(upperKey string, now time.Time, dataPoints *map[string]data.Points, m map[string]interface{}) {
	for k, v := range m {
		key := k
		if upperKey != "" {
			key = fmt.Sprintf("%s.%s", upperKey, k)
		}
		var value float64
		var err error
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
			if err != nil {
				continue
			}
		case []interface{}:
			jsonArrayAppendData(key, now, dataPoints, v.([]interface{}))
		case map[string]interface{}:
			jsonSubMapAppendData(key, now, dataPoints, v.(map[string]interface{}))
		default:
			continue
		}

		(*dataPoints)[key] = append((*dataPoints)[key], data.Point{
			Timestamp: now,
			Value:     value,
		})
	}
}

func jsonArrayAppendData(upperKey string, now time.Time, dataPoints *map[string]data.Points, m []interface{}) {
	for _, v := range m {
		var value float64
		var err error
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
			if err != nil {
				continue
			}
		case []interface{}:
			jsonArrayAppendData(upperKey, now, dataPoints, v.([]interface{}))
		case map[string]interface{}:
			jsonSubMapAppendData(upperKey, now, dataPoints, v.(map[string]interface{}))
		default:
			continue
		}

		(*dataPoints)[upperKey] = append((*dataPoints)[upperKey], data.Point{
			Timestamp: now,
			Value:     value,
		})
	}
}

func JsonDataToResult(raw []byte) (*Result, error) {
	var m map[string]interface{}
	err := json.Unmarshal(raw, &m)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	dataPoints := make(map[string]data.Points, 0)
	jsonSubMapAppendData("", now, &dataPoints, m)
	return &Result{DataPoints: dataPoints, Err: nil}, nil
}
