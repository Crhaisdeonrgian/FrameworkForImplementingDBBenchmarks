package loader

import (
	"FrameworkForImplementingDBBenchmarks/data"
	"FrameworkForImplementingDBBenchmarks/linker"
	"FrameworkForImplementingDBBenchmarks/test"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Loader interface {
	Workload() error
	BackgroundWorkload() error
}
type LoadParams struct {
	db          *sql.DB
	ContextTime time.Duration
	TestTime    time.Duration
	Period      time.Duration
	Collector   data.LatencyCollector
}

func newDBLoader(options linker.DockerOptions, collector data.LatencyCollector) Loader {
	var loader Loader
	params := LoadParams{
		db:          options.DB,
		ContextTime: 15 * time.Second,
		TestTime:    10 * time.Minute,
		Period:      2 * time.Second,
		Collector:   collector,
	}
	loader = params
	return loader
}

func (lp LoadParams) Workload() error {
	var err error
	var done = make(chan struct{})
	var testEnded = make(chan struct{})
	mediumTicker := time.NewTicker(lp.Period)
	testStart := time.Now()
	go func(chan struct{}, chan struct{}) {
		for {
			select {
			case <-done:
				close(testEnded)
				//TEARDOWN STAGE
				return
			case <-mediumTicker.C:
				go func(chan struct{}) {
					start := time.Now()
					Query(lp.db, test.MediumRead)
					select {
					case <-testEnded:
					default:
						// chan is open so test is running
						d := time.Since(start).Milliseconds()
						lp.Collector.Collect(d, testStart)
						log.Println("MediumQuery duration: ", d)
					}
				}(testEnded)
			}
		}
	}(done, testEnded)
	time.Sleep(lp.TestTime)
	mediumTicker.Stop()
	done <- struct{}{}
	return err
}
func (lp LoadParams) BackgroundWorkload() error {
	var err error
	var done = make(chan struct{})
	hardTicker := time.NewTicker(lp.Period * 3)
	go func(chan struct{}) {
		for {
			select {
			case <-done:
				return
			case <-hardTicker.C:
				go func() {
					err = QueryWithTimeout(lp.db, test.SlowRead, lp.ContextTime)
					if err != nil {
						log.Println("loader error ", err)
					}
					fmt.Println("hard query done")
				}()
			}
		}
	}(done)
	time.Sleep(lp.TestTime)
	hardTicker.Stop()
	done <- struct{}{}
	return err
}

var fakeRows *sql.Rows
var fakeResult sql.Result

func Query(dbStd *sql.DB, query string) error {
	var err error
	fakeRows, err = dbStd.Query(query)
	if err != nil {
		log.Fatal("got error in query execution: ", err)
	}
	return err
}
func QueryWithTimeout(dbStd *sql.DB, query string, timeout time.Duration) error {
	var err error
	queryCtx, queryCancel := context.WithTimeout(context.Background(), timeout)
	defer queryCancel()
	fakeRows, err = dbStd.QueryContext(queryCtx, query)
	if err != nil && err != context.DeadlineExceeded {
		log.Fatal("got error in query with timeout execution: ", err)
	}
	fmt.Println("query with timeout execution done")
	return err
}
func ExecuteQuery(dbStd *sql.DB, query string) error {
	var err error
	fakeResult, err = dbStd.Exec(query)
	if err != nil {
		log.Fatal("got error in query execution: ", err)
	}
	return err
}
func ExecuteQueryWithTimeout(dbStd *sql.DB, query string, timeout time.Duration) error {
	var err error
	queryCtx, queryCancel := context.WithTimeout(context.Background(), timeout)
	defer queryCancel()
	fakeResult, err = dbStd.ExecContext(queryCtx, query)
	if err != nil && err != context.DeadlineExceeded {
		log.Fatal("got error in query with timeout execution: ", err)
	}
	fmt.Println("query with timeout execution done")
	return err
}
