package test

import (
	"FrameworkForImplementingDBBenchmarks/data"
	"FrameworkForImplementingDBBenchmarks/linker"
	"FrameworkForImplementingDBBenchmarks/loader"
	"FrameworkForImplementingDBBenchmarks/ticker"
	"database/sql"
	"fmt"
	"github.com/KyleBanks/dockerstats"
	"log"
	"os"
	"testing"
	"time"
)

const (
	FilePath   string = "/Users/igorvozhga/DIPLOMA/"
	MountPoint string = "/Users/igorvozhga/DIPLOMA/mountDir:/var/lib/mysql"
)

type test struct {
}

var eo = linker.EnvOptions{
	FilePath:    FilePath,
	MountPoints: []string{MountPoint},
}
var bo = linker.DBOptions{
	User:     "root",
	Password: "secret",
}
var options = linker.DockerOptions{Account: bo,
	Folders:    eo,
	Repository: "mysql",
	Tag:        "5.6",
	Env:        []string{"MYSQL_ROOT_PASSWORD=secret"},
}
var latency = data.Latency{
	Name:       "Latency",
	Value:      make(chan int64, 2000),
	Timestamps: make(chan int64, 2000),
}
var ds = data.DockerStats{
	Name:       "CPU&Memory",
	Value:      make(chan []dockerstats.Stats, 2000),
	Timestamps: make(chan int64, 2000),
}
var done = make(chan struct{})
var testEnded = make(chan struct{})

var err error

func setup() (*sql.DB, linker.Config) {
	//options - параметры подключения к среде
	//dbStd - указатель на подключение к базе данных
	//config - конфигурация подключения к контейнеру
	dbStd, config, err := options.Link()
	if err != nil {
		log.Println("error in linker ", err)
	}
	return dbStd, config
}

func dflt(db *sql.DB) {
	err = loader.Query(db, MediumRead)
	if err != nil {
		log.Println("error in loader ", err)
	}
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
	latencyMD := latency.SetMetricData()
	CPUMD := ds.SetCPUMetricData()
	MemoryMD := ds.SetMemoryMetricData()
	collected := data.Collected{latency,
		ds}
	collected.CloseAll()
	latencyMD.Output()
	CPUMD.Output()
	MemoryMD.Output()
	if err := config.Pool.Purge(config.Container); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(0)
}

var threadPool = make(chan struct{}, 10)

func TestLoad(t *testing.T) {

	//SETUP STAGE
	dbStd, config := setup()

	//TEST STAGE
	hardTicker := time.NewTicker(5 * time.Second)
	mediumTicker := time.NewTicker(2 * time.Second)
	statTicker := time.NewTicker(1 * time.Second)
	testStart := time.Now()

	go func(chan struct{}, chan struct{}) {
		for {
			select {
			case <-statTicker.C:
				err = ds.Collect(testStart)
				if err != nil {
					log.Println("error in collector: ", err)
				}
			case <-done:
				close(testEnded)
				//TEARDOWN STAGE
				teardown(config)
				return
			case <-hardTicker.C:
				go func() {
					err = loader.QueryWithTimeout(dbStd, SlowRead, 15*time.Second)
					if err != nil {
						log.Println("loader error ", err)
					}
					fmt.Println("hard query done")
				}()
			case <-mediumTicker.C:
				go func(chan struct{}) {
					start := time.Now()
					dflt(dbStd)
					select {
					case _, isOpen := <-testEnded:
						if isOpen {
							// no one isn't putting anything
							log.Fatal("how u did it??")
						} else {
							// chan closed doing nothing
						}
					default:
						// chan is open so test is running
						d := time.Since(start).Milliseconds()
						latency.Collect(d, testStart)
						log.Println("MediumQuery duration: ", d)
					}
				}(testEnded)
			}
		}
	}(done, testEnded)
	time.Sleep(180 * time.Second)
	hardTicker.Stop()
	mediumTicker.Stop()
	statTicker.Stop()
	done <- struct{}{}

}
func TestRealLoad(t *testing.T) {

	//SETUP STAGE
	dbStd, config := setup()

	//TEST STAGE
	hardTicker := time.NewTicker(5 * time.Second)
	mediumTicker := time.NewTicker(2 * time.Second)
	statTicker := time.NewTicker(1 * time.Second)
	realTicker := ticker.NewRandomTicker(2, 1*time.Minute)
	testStart := time.Now()

	go func(chan struct{}, chan struct{}) {
		for {
			select {
			case <-statTicker.C:
				err = ds.Collect(testStart)
				if err != nil {
					log.Println("error in collector: ", err)
				}
			case <-done:

				close(threadPool)
				close(testEnded)
				//TEARDOWN STAGE
				teardown(config)
				return

			case <-realTicker.C:
				go func(chan struct{}, chan struct{}) {
					threadPool <- struct{}{}
					start := time.Now()
					//time.Sleep(5*time.Second)

					//DEFAULT STAGE
					dflt(dbStd)

					select {
					case _, isOpen := <-testEnded:
						if isOpen {
							// no one isn't putting anything
							log.Fatal("how u did it??")
						} else {
							// chan closed doing nothing
						}
					default:
						// chan is open so test is running
						d := time.Since(start).Milliseconds()
						log.Println("MediumQuery duration: ", d)
						latency.Collect(d, testStart)
						<-threadPool
					}
				}(testEnded, threadPool)
			}
		}
	}(done, testEnded)
	time.Sleep(180 * time.Second)
	hardTicker.Stop()
	mediumTicker.Stop()
	statTicker.Stop()
	done <- struct{}{}

}
