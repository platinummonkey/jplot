package data

import (
	"sort"
	"sync"
	"time"
)

type Point struct {
	Timestamp time.Time
	Value     float64
}

func (p *Point) Less(o Point) bool {
	return p.Timestamp.Before(o.Timestamp)
}

func (p *Point) Equal(o Point) bool {
	return p.Timestamp.Equal(o.Timestamp) && p.Value == o.Value
}

type Points []Point

func (p Points) Len() int {
	return len(p)
}

func (p Points) Less(i, j int) bool {
	return p[i].Less(p[j])
}

func (p Points) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Points) XYValues() ([]time.Time, []float64) {
	xVals := make([]time.Time, 0, p.Len())
	yVals := make([]float64, 0, p.Len())
	for _, v := range p {
		xVals = append(xVals, v.Timestamp)
		yVals = append(yVals, v.Value)
	}
	return xVals, yVals
}

type SortedPointSet struct {
	Size    int
	timeMap map[time.Time]Point
	last    Point
}

func (s *SortedPointSet) Add(p Point) {
	point, found := s.timeMap[p.Timestamp]
	if !found {
		if len(s.timeMap) < s.Size {
			s.timeMap[p.Timestamp] = p
		} else {
			if s.last.Timestamp.Unix() <= 0 { // initialization case
				s.last = p
				s.timeMap[p.Timestamp] = p
			} else if s.last.Timestamp.Before(p.Timestamp) {
				// last one can be dropped
				delete(s.timeMap, s.last.Timestamp)
				s.timeMap[p.Timestamp] = p
				// find the oldest
				oldest := p.Timestamp
				for k, _ := range s.timeMap {
					if k.Before(oldest) {
						oldest = k
					}
				}
				s.last = s.timeMap[oldest]
			}
		}
	} else {
		// update value
		point.Value = p.Value
	}
}

func (s *SortedPointSet) Points() Points {
	points := make(Points, 0)
	for _, v := range s.timeMap {
		points = append(points, v)
	}
	sort.Sort(points)
	return points
}

func (s *SortedPointSet) Len() int {
	return len(s.timeMap)
}

type DataSet struct {
	// Size is the number of data point to store per metric.
	Size              int
	ExpectedFrequency time.Duration

	points map[string]SortedPointSet
	last   map[string]Point
	mu     sync.Mutex
}

func (ds *DataSet) Push(name string, ts time.Time, value float64, counter bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if counter {
		var diff float64
		if last := ds.last[name]; last.Value > 0 {
			diff = value - last.Value
		}
		ds.last[name] = Point{Timestamp: ts, Value: value}
		value = diff
	}
	newPoint := Point{Timestamp: ts, Value: value}
	ds.pushPoint(name, newPoint)
}

func (ds *DataSet) PushPoints(name string, data Points, counter bool) {
	sort.Sort(&data)
	ds.mu.Lock()
	defer ds.mu.Unlock()
	for _, p := range data {
		if counter {
			var diff float64
			if last := ds.last[name]; last.Value > 0 {
				diff = p.Value - last.Value
			}
			ds.last[name] = Point{Timestamp: p.Timestamp, Value: p.Value}
			p.Value = diff
		}
		ds.pushPoint(name, p)
	}
}

func (ds *DataSet) Get(name string) Points {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	sps := ds.getLocked(name)
	return sps.Points()
}

func (ds *DataSet) pushPoint(name string, p Point) {
	d := ds.getLocked(name)
	d.Add(p)
	ds.points[name] = d
}

func (ds *DataSet) getLocked(name string) SortedPointSet {
	if ds.points == nil {
		ds.points = make(map[string]SortedPointSet, 0)
		ds.last = make(map[string]Point)
	}
	d, found := ds.points[name]
	if !found {
		d = SortedPointSet{
			Size:    ds.Size,
			timeMap: make(map[time.Time]Point, 0),
		}
		ds.points[name] = d
	}
	return d
}
