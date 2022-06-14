package loader

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Loader interface {
	Query(dbStd *sql.DB, query string) error
	QueryWithTimeout(dbStd *sql.DB, query string, timeout time.Duration) error
	ExecuteQuery(dbStd *sql.DB, query string) error
	ExecuteQueryWithTimeout(dbStd *sql.DB, query string, timeout time.Duration) error
}
var fakeRows *sql.Rows
var fakeResult sql.Result


func Query(dbStd * sql.DB, query string) error{
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
func ExecuteQuery(dbStd * sql.DB, query string) error{
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

