package cmssynchronizer

import (
	"log"
	"os"

	"github.com/mikespook/gearman-go/worker"
)

const htDocs string = "/var/www/vhosts/content.dnalc.org/htdocs"

var sites = map[string]string{
	"dnalc":           "dnalc.cshl.edu",
	"dnabarcoding101": "dnabarcoding101.org",
	"learnaboutsma":   "learnaboutsma.org",
	"maizecode":       "maizecode.org",
	"dnaftb":          "dnaftb.org",
	"summercamps":     "summercamps.dnalc.org",
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

//Ping - will return "Pong"
func Ping(job worker.Job) ([]byte, error) {
	//data := string(job.Data())
	log.Println("Got Ping request..")

	return []byte("Pong"), nil
}

//Synchronize - will invoke svn update on the given path
func Synchronize(job worker.Job) ([]byte, error) {
	data := string(job.Data())
	log.Println("Got ", data)

	/*var outbuf, errbuf bytes.Buffer
	var sitedir string
	var exists bool
	if sitedir, exists = sites[site]; !exists {
		fmt.Println("Not found: ", sitedir)
		return []byte("Error: Invalid request"), nil
	}

	(fullPath := virtualDir + "/" + sitedir
	//fullPath := "/Users/cornel/work/" + sitedir

	if exists, _ = pathExists(fullPath); !exists {
		return []byte(fmt.Sprintf("Not found root directory for %s!", site)), nil
	}*/

	out := ""

	return []byte(out), nil
}
