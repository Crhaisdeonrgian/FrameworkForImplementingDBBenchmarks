package test

import (
	"context"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

/*
Fast read is about 10ms
Medium read is about 5s
Slow read is about 3min
*/
const (
	FastRead   string = "SELECT * FROM tbl WHERE col_a"
	MediumRead        = "SELECT * FROM tbl WHERE id<110000 ORDER BY col_b DESC, col_a ASC"
	SlowRead          = "SELECT * FROM tbl first JOIN tbl second ON second.id<5 " +
		"WHERE first.col_a LIKE '%a%' ORDER BY first.col_b DESC, first.col_a ASC"
	CustomRead  = "SELECT * FROM tbl WHERE id<?"
	CustomWrite = "INSERT INTO tbl(col_a, col_b, col_c, col_d) VALUES (?,?,?,?)"
	DropTable   = "DROP TABLE tbl"
	CreateTable = "CREATE TABLE IF NOT EXISTS tbl  (id int AUTO_INCREMENT PRIMARY KEY, a_col nvarchar(1025), b_col nvarchar(1025), c_col nvarchar(1025), d_col nvarchar(1025) )"
)
const CPUCount = 1
const Memory = 1024 * 1024 * 1024
const driverName = "mysql"
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890йцукенгшщзхъфывапролджэёячсмитьбюЙЦУКЕНГШЩЗХЪФЫВАПРОЛДЖЭЁЯЧСМИТЬБЮ"

// nolint:gochecknoglobals
var systemdb *sql.DB // the connection to the mysql 'system' database
// nolint:gochecknoglobals
var sqlConfig *mysql.Config // the mysql container and config for connecting to other databases
// nolint:gochecknoglobals
var testMu *sync.Mutex // controls access to sqlConfig

func RandStringBytes() string {
	b := make([]byte, 1024)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%194]
	}
	return string(b)
}

func CreateDatabaseTable(db *sql.DB) {
	var err error
	_, err = db.Exec(DropTable)
	_, err = db.Exec(CreateTable)
	if err != nil {
		log.Fatal(err)
	}
}
func FillDataBaseTable(db *sql.DB) {
	var tx *sql.Tx
	var err error
	for i := 0; i < 10000; i++ {
		currentctx, currentcancel := context.WithTimeout(context.Background(), 12*time.Hour)
		defer currentcancel()
		tx, err = db.BeginTx(currentctx, nil)
		if err != nil {
			log.Fatal(err)
		}
		_, err = tx.ExecContext(currentctx, CustomWrite, RandStringBytes(), RandStringBytes(), RandStringBytes(), RandStringBytes())
		if err != nil {
			log.Fatal(err, tx.Rollback())
		}
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
		/*if i%100 == 0 {
			fmt.Println(strconv.Itoa(i))
		}*/
	}
}
func connectToDB() *sql.DB {
	var err error
	var dbStd *sql.DB
	testMu.Lock()
	benchTestConfig := sqlConfig
	testMu.Unlock()
	if err != nil {
		log.Fatal(err)
	}
	dbStd, err = sql.Open(driverName, benchTestConfig.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	return dbStd
}

func ShowDatabases() {
	var err error
	var rows *sql.Rows
	rows, err = systemdb.Query("show databases")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	second_params := make([]string, 0)
	for rows.Next() {
		var second string
		if err := rows.Scan(&second); err != nil {
			log.Fatal(err)
		}
		second_params = append(second_params, second)
	}
	log.Println("all the bases")
	log.Println(strings.Join(second_params, " "))
}
func CheckRows(db *sql.DB) int {
	var err error
	var rows *sql.Rows
	queryctx, querycancel := context.WithTimeout(context.Background(), 100000000*time.Millisecond)
	defer querycancel()
	rows, err = db.QueryContext(queryctx, "select count(*) from abobd")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var first int
		if err := rows.Scan(&first); err != nil {
			log.Fatal(err)
		}
		return first
	}
	return 0
}
