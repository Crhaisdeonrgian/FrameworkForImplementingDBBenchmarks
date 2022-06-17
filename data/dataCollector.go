package data

import (
	"FrameworkForImplementingDBBenchmarks/loader"
	"github.com/KyleBanks/dockerstats"
	"log"
	"time"
)

type LatencyCollector interface {
	Collect(latency int64, testStart time.Time)
	GetData() MetricData
	Close()
}
type DockerStatsCollector interface {
	Collect(testStart time.Time) error
	GetData() MetricData
	GetStats() DockerStats
	Close()
}

func newLatencyCollector() LatencyCollector {
	var latency = Latency{
		Name:       "Latency",
		Value:      make(chan int64, 2000),
		Timestamps: make(chan int64, 2000),
	}
	var collector LatencyCollector
	collector = latency
	return collector
}

func (l Latency) GetData() MetricData {
	m := MetricData{}
	m.Timestamps, m.Data = l.getStringData()
	m.XValues, m.YValues = l.preparePoints()
	return m
}
func (ds DockerStats) GetStats() DockerStats {
	return ds
}
func (ds DockerStats) GetData() MetricData {
	m := MetricData{}
	m.Timestamps, m.Data = ds.getStringData()
	m.XValues, m.YValues = ds.prepareMemoryPoints()
	return m
}

func newDockerStatsCollector() DockerStatsCollector {
	var collector DockerStatsCollector
	var ds = DockerStats{
		Name:       "CPU&Memory",
		Value:      make(chan []dockerstats.Stats, 2000),
		Timestamps: make(chan int64, 2000),
	}
	collector = ds
	return collector
}

func (l Latency) Collect(latency int64, testStart time.Time) {
	l.Timestamps <- time.Since(testStart).Milliseconds()
	l.Value <- latency
}

func (ds DockerStats) StartCollecting(params loader.LoadParams) {
	statTicker := time.NewTicker(1 * time.Second)
	var done = make(chan struct{})
	testStart := time.Now()
	go func(chan struct{}) {
		for {
			select {
			case <-statTicker.C:
				err := ds.Collect(testStart)
				if err != nil {
					log.Println("error in collector: ", err)
				}
			case <-done:
				//TEARDOWN STAGE
				return
			}
		}
	}(done)
	time.Sleep(params.TestTime)
	statTicker.Stop()
}

func (ds DockerStats) Collect(testStart time.Time) error {
	s, err := dockerstats.Current()
	if err != nil {
		log.Println("Unable to get stats ", err)
	}
	ds.Timestamps <- time.Since(testStart).Milliseconds()
	ds.Value <- s
	return err
}
func (l Latency) Close() {
	close(l.Value)
}
func (ds DockerStats) Close() {
	close(ds.Value)
}
