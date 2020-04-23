// +build linux
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"gearmanworkers/cmssynchronizer"

	"github.com/mikespook/gearman-go/worker"
	"github.com/mikespook/golib/signal"
)

var dbh *sql.DB

func init() {

	if "" == os.Getenv("GEARMAN_SERVERS") {
		log.Fatalln("GEARMAN_SERVERS not set")
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = os.Getenv("USER")
	}

	dbPass := os.Getenv("DB_PASS")

	dbName := os.Getenv("DB_DATABASE")
	if "" == dbName {
		log.Fatalln("DB_DATABASE not set")
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	// default collation is 'utf8mb4_general_ci'
	db, err := sql.Open("mysql", dbUser+":"+dbPass+"@tcp("+dbHost+")/"+dbName)
	if err != nil {
		fmt.Println(dbUser + ":XXX" + "@tcp(" + dbHost + ")/" + dbName)
		panic(err.Error())
	}

	// Open doesn't open a connection, so let's Ping() our db
	err = db.Ping()
	if err != nil {
		fmt.Println(dbUser + ":XXX" + "@tcp(" + dbHost + ")/" + dbName)
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	dbh = db
}

func main() {

	defer dbh.Close()
	defer log.Println("Shutdown complete!")
	w := worker.New(worker.OneByOne)
	defer w.Close()

	/*atomGetter := cmssynchronizer.Atoms{DB: dbh}
	atom, err := atomGetter.GetByID(17136)
	if err != nil {
		panic(err.Error())
	}
	//fmt.Printf("atom = %+v\n", atom)
	fmt.Printf("[%s] %s\n", *atom.ID, *atom.Name)
	for _, ad := range atom.Downloads {
		fmt.Println(" Type: ", *ad.Type)
		fmt.Println(" Label:", *ad.Label)
		fmt.Println(" Path: ", *ad.Path)
		fmt.Println("")
	}
	*/

	w.ErrorHandler = func(e error) {
		log.Println(e)
		if opErr, ok := e.(*net.OpError); ok {
			if !opErr.Temporary() {
				proc, err := os.FindProcess(os.Getpid())
				if err != nil {
					log.Println(err)
				}
				if err := proc.Signal(os.Interrupt); err != nil {
					log.Println(err)
				}
			}
		}
	}

	w.JobHandler = func(job worker.Job) error {
		log.Printf("Data=%s\n", job.Data())
		return nil
	}
	gServers := strings.Split(os.Getenv("GEARMAN_SERVERS"), ",")
	for _, srv := range gServers {
		log.Printf("++ adding gearman server %s\n", srv)
		w.AddServer("tcp4", srv)
	}

	w.AddFunc("SynchAtomFiles", cmssynchronizer.SynchAtomFiles, 10)
	w.AddFunc("FixAtomPems", cmssynchronizer.FixAtomPems, 10)
	//w.AddFunc("Ping", cmssynchronizer.Ping, 30)

	if err := w.Ready(); err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("waiting for work..")

	go w.Work()
	signal.Bind(os.Interrupt, func() uint { return signal.BreakExit })
	signal.Wait()
}
