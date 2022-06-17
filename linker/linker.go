package linker

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const (
	FilePath   string = "/Users/igorvozhga/DIPLOMA/"
	MountPoint string = "/Users/igorvozhga/DIPLOMA/mountDir:/var/lib/mysql"
	CPUCount          = 1
	Memory            = 1024 * 1024 * 1024
	driverName        = "mysql"
)

type DBOptions struct {
	user     string
	password string
}
type EnvOptions struct {
	filePath    string   //Путь к папке для сохранения результатов
	mountPoints []string //Путь к папке, в которую монтируется том контейнера
}
type DockerOptions struct {
	account    DBOptions
	folders    EnvOptions
	repository string
	tag        string
	env        []string
	DB         *sql.DB
	config     Config
}

type Connections []Linker

type Config struct {
	pool      *dockertest.Pool
	container *dockertest.Resource
}

type Linker interface {
	Link() error
	Close()
	//GetData()
}

func newDockerLinker() Linker {
	file := flag.String("saveto", FilePath, "specify where to save collected data")
	mount := flag.String("mountPoint", MountPoint, "specify where to mount volume")
	flag.Parse()
	var dockerLinker Linker
	var eo = EnvOptions{
		filePath:    *file,
		mountPoints: []string{*mount},
	}
	var bo = DBOptions{
		user:     "root",
		password: "secret",
	}
	var options = DockerOptions{account: bo,
		folders:    eo,
		repository: "mysql",
		tag:        "5.6",
		env:        []string{"MYSQL_ROOT_PASSWORD=secret"},
	}
	dockerLinker = options
	return dockerLinker
}

func (c Connections) SetConnections() error {
	for _, conn := range c {
		err := conn.Link()
		if err != nil {
			log.Fatal("Error in linking connection: ", err)
			return err
		}
	}
	return nil
}

//Implementation

//newDockerLinker()
//closeDocker()
//closeDB()

func (myOptions DockerOptions) Link() error {
	//СПРЯТАЛЬ
	var dockerPool *dockertest.Pool                     // the connection to docker
	var systemdb *sql.DB                                // the connection to the mysql 'system' database
	var sqlConfig *mysql.Config                         // the mysql container and config for connecting to other databases
	var testMu *sync.Mutex                              // controls access to sqlConfig
	_ = mysql.SetLogger(log.New(ioutil.Discard, "", 0)) // silence mysql logger
	testMu = &sync.Mutex{}

	var err error
	dockerPool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}
	dockerPool.MaxWait = time.Minute * 2

	runOptions := dockertest.RunOptions{
		Repository: myOptions.repository,
		Tag:        myOptions.tag,
		Env:        myOptions.env,
		Mounts:     myOptions.folders.mountPoints,
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
		User:                 myOptions.account.user,
		Passwd:               myOptions.account.password,
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
	myOptions.config = Config{pool: dockerPool, container: mysqlContainer}
	myOptions.DB = dbStd
	return err
}

func (myOptions DockerOptions) Close() {
	if err := myOptions.config.pool.Purge(myOptions.config.container); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	os.Exit(0)
}
