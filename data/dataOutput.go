package data

import (
	"github.com/wcharczuk/go-chart"
	"log"
	"os"
)

type Outputer interface {
	Output()
}

func (m MetricData) Output(){
	//Разнести по разным флагам запуска
	m.WriteToFile()
	m.DrawPlot()
}
type Results []Outputer


func (m MetricData) WriteToFile() {
	file, err := os.Create(m.Name + "Result" + ".csv")
	if err != nil {
		log.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("Unable to close file:", err)
			os.Exit(1)
		}
	}(file)
	_, err = file.WriteString("time" + "," + m.Name + "\n")
	if err != nil {
		log.Println("Unable to write to file:", err)
	}
	for time , row := range m.Data {
		_, err = file.WriteString(m.Timestamps[time] + "," + row +"\n")
		if err != nil {
			log.Println("Unable to write to file:", err)
		}
	}
}



func (m MetricData) DrawPlot() {

	fd, err := os.Create( m.Name + "Plot.png")
	if err != nil {
		log.Fatal("cannot open file ", err)
	}

	plot := chart.Chart{
		Title: m.Name,
		XAxis: chart.XAxis{Name: "t,sec", Style: chart.Style{Show: true}},
		YAxis: chart.YAxis{Name: "Query duration,sec", Style: chart.Style{Show: true}},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{

					StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
					FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
					Show:        true,
					Padding:     chart.DefaultBackgroundPadding,
				},
				XValues: m.XValues,
				YValues: m.YValues,
			},
		},
	}
	err = plot.Render(chart.PNG, fd)
	if err != nil {
		log.Fatal("got error while rendering plot: ", err)
	}
	err = fd.Close()
	if err != nil {
		log.Fatal("cannot close file ", err)
	}
}
