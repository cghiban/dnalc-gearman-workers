// +build linux
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"gearmanworkers/cmssynchronizer"

	"github.com/mikespook/gearman-go/worker"
	"github.com/mikespook/golib/signal"
)

func init() {

	if "" == os.Getenv("CONTENT_HTDOCS") {
		log.Fatalln("CONTENT_HTDOCS not set")
	}

	if "" == os.Getenv("GEARMAN_SERVERS") {
		log.Fatalln("GEARMAN_SERVERS not set")
	}
}

func main() {

	defer log.Println("Shutdown complete!")
	w := worker.New(worker.OneByOne)
	defer w.Close()

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

	w.AddFunc("SynchAtomFiles", cmssynchronizer.SynchAtomFiles, 15)
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
