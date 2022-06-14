package data

import (
	"fmt"
	"github.com/KyleBanks/dockerstats"
	"log"
	"strconv"
	"strings"
)


type DockerStats struct {
	Name string
	Value chan []dockerstats.Stats
	Timestamps chan int64
}
//CРАЗУ ПРИ СБОРЕ
type MetricData struct {
	Name       string
	Data       []string
	Timestamps []string
	XValues    []float64
	YValues []float64
}
type Latency struct {
	Name string
	Value chan int64
	Timestamps chan int64
}
func (l Latency) SetMetricData() MetricData {
	m := MetricData{}
	m.Timestamps, m.Data = l.getStringData()
	m.XValues, m.YValues = l.preparePoints()
	return m
}
func (ds DockerStats) SetMemoryMetricData() MetricData {
	m := MetricData{}
	m.Timestamps, m.Data = ds.getStringData()
	m.XValues, m.YValues = ds.prepareMemoryPoints()
	return m
}
func (ds DockerStats) SetCPUMetricData() MetricData {
	m := MetricData{}
	m.Timestamps, m.Data = ds.getStringData()
	m.XValues, m.YValues = ds.prepareCPUPoints()
	return m
}

func (l Latency) getStringData() ([]string, []string){
	var data []string
	var timestamps []string
	for value := range l.Value{
		data = append(data, fmt.Sprint(value))
	}
	for stamp := range l.Timestamps{
		timestamps = append(timestamps, fmt.Sprint(stamp))
	}
	return timestamps, data
}
func (ds DockerStats) getStringData() ([]string,[]string)  {
	var data []string
	var timestamps []string
	for values := range ds.Value{
		for _, value := range values {
			data = append(data, value.String())
		}
	}
	for stamp := range ds.Timestamps{
		timestamps = append(timestamps, fmt.Sprint(stamp))
	}
	return timestamps,data
}

func (l Latency) preparePoints() ([]float64, []float64) {
	var XValues []float64
	var YValues []float64
	for x := range l.Timestamps {
		YValues = append(YValues, float64(x))
	}
	for y := range l.Value {
		YValues = append(YValues, float64(y))
	}
	return XValues,YValues
}
func (ds DockerStats) prepareMemoryPoints()([]float64, []float64)  {
	var XValues []float64
	var YValues []float64
	for x := range ds.Timestamps {
		XValues = append(XValues, float64(x))
	}
	for ys := range ds.Value {
		for _, y := range ys{
			YValues = append(YValues, parseMemoryStats(y.Memory) )
		}
	}
	return XValues, YValues
}

func (ds DockerStats) prepareCPUPoints()([]float64, []float64)  {
	var XValues []float64
	var YValues []float64
	for x := range ds.Timestamps {
		XValues = append(XValues, float64(x))
	}
	for ys := range ds.Value {
		for _, y := range ys{
			YValues = append(YValues, parseCPUStats(y) )
		}
	}
	return XValues, YValues
}

func parseMemoryStats(mem dockerstats.MemoryStats) float64{
	percent := mem.Percent
	number := strings.TrimSuffix(percent, "%")
	i, err := strconv.ParseInt(number, 10, 64)
	if err!=nil{
		log.Println("Error in parsing memory", err)
	}

	return float64(i)
}
func parseCPUStats(mem dockerstats.Stats) float64{
	number := mem.CPU
	i, err := strconv.ParseInt(number, 10, 64)
	if err!=nil{
		log.Println("Error in parsing memory", err)
	}
	return float64(i)
}
