package window

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"

	humanize "github.com/dustin/go-humanize"
	"github.com/rs/jplot/data"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	//"github.com/wcharczuk/go-chart/seq"
)

func Init() {
	chart.DefaultBackgroundColor = chart.ColorTransparent
	chart.DefaultCanvasColor = chart.ColorTransparent
	chart.DefaultTextColor = drawing.Color{R: 180, G: 180, B: 180, A: 255}
	chart.DefaultAxisColor = drawing.Color{R: 180, G: 180, B: 180, A: 255}
	chart.DefaultAnnotationFillColor = chart.ColorBlack.WithAlpha(200)
}

func Clear() {
	print("\033\133\110\033\133\062\112") // clear screen
	print("\033]1337;CursorShape=1\007")  // set cursor to vertical bar
}

func Reset() {
	print("\033\133\061\073\061\110") // move cursor to 0x0
}

// Graph generate a line graph with series.
func Graph(series []chart.Series, markers []chart.GridLine, width, height int) chart.Chart {
	for i, s := range series {
		if s, ok := s.(chart.TimeSeries); ok {
			//s.XValues = seq.Range(0, float64(len(s.YValues)-1))
			c := chart.GetAlternateColor(i + 4)
			s.Style = chart.Style{
				Show:        true,
				StrokeWidth: 2,
				StrokeColor: c,
				FillColor:   c.WithAlpha(20),
				FontSize:    9,
			}
			series[i] = s
			last := chart.LastValueAnnotation(s, SIValueFormater)
			last.Style.FillColor = c
			last.Style.FontColor = TextColor(c)
			last.Style.FontSize = 9
			last.Style.Padding = chart.NewBox(2, 2, 2, 2)
			series = append(series, last)
		}
	}
	graph := chart.Chart{
		Width:  width,
		Height: height,
		Background: chart.Style{
			Padding: chart.NewBox(5, 0, 0, 5),
		},
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: chart.TimeMinuteValueFormatter,
			//ValueFormatter: chart.TimeDateValueFormatter,
		},
		YAxis: chart.YAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: SIValueFormater,
		},
		Series: series,
	}
	if len(markers) > 0 {
		graph.Background.Padding.Bottom = 0 // compensate transparent tick space
		graph.XAxis = chart.XAxis{
			Style: chart.StyleShow(),
			TickStyle: chart.Style{
				StrokeColor: chart.ColorTransparent,
			},
			TickPosition: 10, // hide text with non-existing position
			GridMajorStyle: chart.Style{
				Show:            true,
				StrokeColor:     chart.ColorAlternateGray.WithAlpha(100),
				StrokeWidth:     2.0,
				StrokeDashArray: []float64{2.0, 2.0},
			},
			GridLines: markers,
		}
	}
	graph.Elements = []chart.Renderable{
		Legend(&graph, chart.Style{
			FillColor:   drawing.Color{A: 100},
			FontColor:   chart.ColorWhite,
			StrokeColor: chart.ColorTransparent,
		}),
	}
	return graph
}

func TextColor(bg drawing.Color) drawing.Color {
	var L float64
	for c, f := range map[uint8]float64{bg.R: 0.2126, bg.G: 0.7152, bg.B: 0.0722} {
		c := float64(c) / 255.0
		if c <= 0.03928 {
			c = c / 12.92
		} else {
			c = math.Pow(((c + 0.055) / 1.055), 2.4)
		}
		L += c * f
	}
	if L > 0.179 {
		return chart.ColorBlack
	}
	return chart.ColorWhite
}

func SIValueFormater(v interface{}) string {
	var value float64
	var prefix string
	switch v.(type) {
	case float64:
		value, prefix = humanize.ComputeSI(v.(float64))
	case data.Point:
		value, prefix = humanize.ComputeSI((v.(data.Point)).Value)
	}
	value = float64(int(value*100)) / 100
	return humanize.Ftoa(value) + " " + prefix
}

// PrintGraphs generates a single PNG with graphs stacked and print it to iTerm2.
func PrintGraphs(graphs []chart.Chart) {
	var width, height int
	for _, graph := range graphs {
		if graph.Width > width {
			width = graph.Width
		}
		height += graph.Height
	}
	Reset()
	canvas := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	var top int
	for _, graph := range graphs {
		iw := &chart.ImageWriter{}
		graph.Render(chart.PNG, iw)
		img, _ := iw.Image()
		r := image.Rectangle{image.Point{0, top}, image.Point{width, top + graph.Height}}
		top += graph.Height
		draw.Draw(canvas, r, img, image.Point{0, 0}, draw.Src)
	}
	var b bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &b)
	defer enc.Close()
	png.Encode(enc, canvas)
	fmt.Printf("\033]1337;File=preserveAspectRatio=1;inline=1:%s\007", b.Bytes())
}

func Render(specs []data.GraphSpec, ds *data.DataSet, width, height int) {
	graphs := make([]chart.Chart, 0, len(specs))
	for _, gs := range specs {
		series := []chart.Series{}
		markers := []chart.GridLine{}
		for _, f := range gs.Fields {
			vals := ds.Get(f.ID)
			if f.Marker {
				for i, v := range vals {
					if v.Value > 0 {
						markers = append(markers, chart.GridLine{Value: float64(i)})
					}
				}
				continue
			}
			xVals, yVals := vals.XYValues()
			series = append(series, chart.TimeSeries{
				Name:    fmt.Sprintf("%s: %s", f.Name, SIValueFormater(yVals[len(yVals)-1])),
				XValues: xVals,
				YValues: yVals,
			})
		}
		graphs = append(graphs, Graph(series, markers, width, height/len(specs)))
	}
	PrintGraphs(graphs)
}
