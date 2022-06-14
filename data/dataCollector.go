package data

import (
	"github.com/KyleBanks/dockerstats"
	"log"
	"time"
)

type Collector interface {
	Collect()
}
type Closer interface {
	Close()
}
func (data Data) Collect() {
	for _ , metric := range data {
		metric.Collect()
	}
}
type Collected []Closer

func (c Collected) CloseAll()  {
	for _, metric := range c{
		metric.Close()
	}
}

type Data []Collector

func (l Latency) Collect(latency int64, testStart time.Time){
	l.Timestamps <- time.Since(testStart).Milliseconds()
	l.Value <- latency
}
func (ds DockerStats) Collect(testStart time.Time) error {
	s, err := dockerstats.Current()
	if err!=nil {
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
