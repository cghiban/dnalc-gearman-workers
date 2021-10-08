package svnupdater

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/mikespook/gearman-go/worker"
)

const virtualDir string = "/var/www/virtuals"

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

//Update - will invoke svn update on the given path
func Update(job worker.Job) ([]byte, error) {
	site := string(job.Data())

	var outbuf, errbuf bytes.Buffer
	var sitedir string
	var exists bool
	if sitedir, exists = sites[site]; !exists {
		fmt.Println("Not found: ", sitedir)
		return []byte("Error: Invalid request"), nil
	}
	log.Println("Updating ", site)

	fullPath := virtualDir + "/" + sitedir
	//fullPath := "/Users/cornel/work/" + sitedir

	if exists, _ = pathExists(fullPath); !exists {
		return []byte(fmt.Sprintf("Not found root directory for %s!", site)), nil
	}

	//cmd := exec.Command("ls", fullPath)
	cmd := exec.Command("/usr/bin/svn", "--username", "cornel", "--password", "xxxxx", "--non-interactive", "update", fullPath)
	//out, err := cmd.CombinedOutput()
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	if err != nil {
		fmt.Println("Err:", err)
	}

	out := outbuf.String()
	if errbuf.Len() > 0 {
		out += "\n#---Err:\n" + errbuf.String()
		out += "\n#---Err:\n" + err.Error()
	}

	return []byte(out), nil
}
