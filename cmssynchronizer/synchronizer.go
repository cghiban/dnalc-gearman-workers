package cmssynchronizer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mikespook/gearman-go/worker"
)

//const htDocs string = "/var/www/vhosts/content.dnalc.org/htdocs"
const htDocs string = "testdata"

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

func computeAtomPath(atomID int) string {
	atomIDStr := strconv.Itoa(atomID)
	subpath := fmt.Sprintf("c%d", atomID/1000)
	return filepath.Join(htDocs, "content", subpath, atomIDStr)
}

//FixAtomPems - will mirror atom files
func FixAtomPems(job worker.Job) ([]byte, error) {
	data := string(job.Data())
	atomID, err := strconv.Atoi(data)
	if err != nil {
		return []byte("Invalid atom id:" + err.Error()), nil
	}
	atomDir := computeAtomPath(atomID)
	log.Println("Got atom", atomID)
	log.Println("  ", atomDir)

	// lets to the walk
	err = filepath.Walk(atomDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			fmt.Printf("visited dir: %q\n", path)
			os.Chmod(path, 0755)
		} else if info.Mode().IsRegular() {
			fmt.Printf("visited regular file: %q\n", path)
			fmt.Printf("  pems: %q\n", info.Mode().Perm())
			if err := os.Chmod(path, 0644); err != nil {
				log.Printf("chmod(%s) = %+v\n", path, err.Error())
			}

		} else {
			fmt.Printf("visited other type of file: %q\n", path)
			fmt.Printf("  pems: %q\n", info.Mode().Perm())
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", atomDir, err)
	}

	out := ""
	return []byte(out), nil
}

//SynchAtomFiles - will mirror atom files
func SynchAtomFiles(job worker.Job) ([]byte, error) {
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
