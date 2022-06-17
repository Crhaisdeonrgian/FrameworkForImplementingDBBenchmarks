package test

import (
	"FrameworkForImplementingDBBenchmarks/data"
	"FrameworkForImplementingDBBenchmarks/linker"
	"database/sql"
	"testing"
)

type test struct {
}

var err error

func setup() {

}

func dflt(db *sql.DB) {
}

func TestExample(t *testing.T) {
	metric := data.MetricData{
		Name:       "Example",
		Data:       []string{"1", "2", "3", "4", "5"},
		Timestamps: []string{"1", "2", "3", "4", "5"},
		XValues:    []float64{1, 2, 3, 4, 5},
		YValues:    []float64{1, 2, 3, 4, 5},
	}
	metric.Output()
}

func teardown(config linker.Config) {

}

var threadPool = make(chan struct{}, 10)

func TestLoad(t *testing.T) {

}
func TestRealLoad(t *testing.T) {

}
