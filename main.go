package main

import (
	"encoding/json"
	"fmt"
	"github.com/wcharczuk/go-chart/v2"
	"os"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"strings"
	"bufio"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mkchart <jsonfile>")
		os.Exit(1)
	}
	if err := chartIt(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
func chartIt(fname string) error {
	data, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	var datapoints storage
	if err := json.Unmarshal(data, &datapoints); err != nil {
		return err
	}
	return render(datapoints, fname)
}

func SplitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

type storage struct {
	Legend   string
	Title    string
	Xvalues  []float64
	XUnit    string
	Yvalues  []float64
	YUnit    string
	Y2values []float64
	Y2Unit   string
}

func render(data storage, fname string) error {
	series := chart.ContinuousSeries{
		Name: data.XUnit,
		Style: chart.Style{
			StrokeColor: chart.GetDefaultColor(0),
			FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
		},
		XValues: data.Xvalues,
		YValues: data.Yvalues,
	}
	graph := chart.Chart{
		Series: []chart.Series{series},
		Title:  data.Title,
		YAxis: chart.YAxis{
			Name: data.YUnit,
			Range: &chart.ContinuousRange{
				Min:        0,
				Max:        100,
				Domain:     1,
				Descending: false,
			},
		},
		TitleStyle: chart.Style{
			FontSize: 8,
			TextWrap: chart.TextWrapWord,
		},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
		DetailsBox(&graph, strings.Split(data.Legend, "\n")),
	}
	fName := fmt.Sprintf("%v.png", fname)
	f, err := os.Create(fName)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := graph.Render(chart.PNG, f); err != nil {
		return err
	}
	fmt.Printf("Rendered file %v\n", fName)
	return nil
}

func DetailsBox(c *chart.Chart, text []string, userDefaults ...chart.Style) chart.Renderable {
	fmt.Printf("rendering texts: %v\n", text)
	return func(r chart.Renderer, box chart.Box, chartDefaults chart.Style) {
		// default style
		defaults := chart.Style{
			FillColor:   drawing.ColorWhite.WithAlpha(50),
			FontColor:   chart.DefaultTextColor,
			FontSize:    8.0,
			StrokeColor: chart.DefaultAxisColor,
			StrokeWidth: chart.DefaultAxisLineWidth,
		}

		var style chart.Style
		if len(userDefaults) > 0 {
			style = userDefaults[0].InheritFrom(chartDefaults.InheritFrom(defaults))
		} else {
			style = chartDefaults.InheritFrom(defaults)
		}

		contentPadding := chart.Box{
			Top:    box.Height(),
			Left:   5,
			Right:  5,
			Bottom: box.Height(),
		}

		contentBox := chart.Box{
			Bottom: box.Height(),
			Left:   50,
		}

		content := chart.Box{
			Top:    contentBox.Bottom - 5,
			Left:   contentBox.Left + contentPadding.Left,
			Right:  contentBox.Left + contentPadding.Left,
			Bottom: contentBox.Bottom - 5,
		}

		style.GetTextOptions().WriteToRenderer(r)

		// measure and add size of text to box height and width
		for _, t := range text {
			if len(t) == 0 {
				continue
			}
			textbox := r.MeasureText(t)
			content.Top -= textbox.Height()
			right := content.Left + textbox.Width()
			content.Right = chart.MaxInt(content.Right, right)
		}

		contentBox = contentBox.Grow(content)
		contentBox.Right = content.Right + contentPadding.Right
		contentBox.Top = content.Top - 5

		// draw the box
		chart.Draw.Box(r, contentBox, style)

		style.GetTextOptions().WriteToRenderer(r)

		// add the text
		ycursor := content.Top
		x := content.Left
		for _, t := range text {
			if len(t) == 0 {
				continue
			}
			textbox := r.MeasureText(t)
			y := ycursor + textbox.Height()
			r.Text(t, x, y)
			ycursor += textbox.Height()
		}
	}
}
