package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"gearmanworkers/svnupdater"

	"github.com/mikespook/gearman-go/worker"
	"github.com/mikespook/golib/signal"
)

//Sum worker function
func Sum(job worker.Job) ([]byte, error) {
	log.Println("Sum: Data=", job.Data())
	//data := []byte(strings.ToUpper(string(job.Data())))
	data := []byte("0")
	return data, nil
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
	w.AddServer("tcp4", "143.48.220.52:4730")

	w.AddFunc("SVNUpdate", svnupdater.Update, 5)
	//w.AddFunc("sum", Sum, 5)
	//w.AddFunc("SysInfo", worker.SysInfo, worker.Unlimited)
	//w.AddFunc("MemInfo", worker.MemInfo, worker.Unlimited)

	if err := w.Ready(); err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("waiting for work..")

	go w.Work()
	signal.Bind(os.Interrupt, func() uint { return signal.BreakExit })
	signal.Wait()
}
