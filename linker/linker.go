package linker

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

const CPUCount = 1
const Memory = 1024*1024*1024
const driverName = "mysql"

type DBOptions struct {
	User     string
	Password string
}
type EnvOptions struct {
	FilePath      string 		//Путь к папке для сохранения результатов
	MountPoints   []string 		//Путь к папке, в которую монтируется том контейнера
}


type Linker interface {
	Link() (*sql.DB, error)
	Close() error
	GetData()
}

type DockerOptions struct {
	Account    DBOptions
	Folders    EnvOptions
	Repository string
	Tag           string
	Env           []string
}

type Connections []Linker

type Config struct {
	Pool *dockertest.Pool
	Container *dockertest.Resource
}

//Банда 4х

/*func newDockerLinker(lalala) (Linker, error) {

}*/

func (c Connections) GetConnections() ([]*sql.DB){
	var dbStds []*sql.DB
	for _, conn := range c {
		dbStd, err := conn.Link()
		if(err!=nil){
			log.Fatal("Error in linking connection: ", err)
			continue
		}
		dbStds = append(dbStds, dbStd)
	}
	return dbStds
}
//Implementation

//newDockerLinker()
//closeDocker()
//closeDB()
func (myOptions DockerOptions) Link() (*sql.DB, Config, error) {
	//СПРЯТАЛЬ
	// nolint:gochecknoglobals
	var dockerPool *dockertest.Pool // the connection to docker
	// nolint:gochecknoglobals
	var systemdb *sql.DB // the connection to the mysql 'system' database
	// nolint:gochecknoglobals
	var sqlConfig *mysql.Config // the mysql container and config for connecting to other databases
	// nolint:gochecknoglobals
	var testMu *sync.Mutex // controls access to sqlConfig
	_ = mysql.SetLogger(log.New(ioutil.Discard, "", 0)) // silence mysql logger
	testMu = &sync.Mutex{}

	var err error
	dockerPool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}
	dockerPool.MaxWait = time.Minute * 2

	runOptions := dockertest.RunOptions{
		Repository: myOptions.Repository,
		Tag:        myOptions.Tag,
		Env:        myOptions.Env,
		Mounts:     myOptions.Folders.MountPoints,
	}
	//TODO: add setup method
	mysqlContainer, err := dockerPool.RunWithOptions(&runOptions, func(hostcfg *docker.HostConfig) {
		hostcfg.CPUCount = CPUCount
		//hostcfg.CPUPercent = 100
		hostcfg.Memory = Memory
	})
	if err != nil {
		log.Fatalf("could not start mysqlContainer: %s", err)
	}
	sqlConfig = &mysql.Config{
		User:                 myOptions.Account.User,
		Passwd:               myOptions.Account.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("localhost:%s", mysqlContainer.GetPort("3306/tcp")),
		DBName:               "MyDB",
		AllowNativePasswords: true,
	}

	if err = dockerPool.Retry(func() error {
		systemdb, err = sql.Open(driverName, sqlConfig.FormatDSN())
		if err != nil {
			return err
		}
		return systemdb.Ping()
	}); err != nil {
		log.Fatal(err)
	}
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

	return dbStd,Config{Pool: dockerPool,
		Container: mysqlContainer}, err
}
