package cmssynchronizer

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/mikespook/gearman-go/worker"
)

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

func computeAtomPath(atomID int) (string, error) {
	atomIDStr := strconv.Itoa(atomID)
	subpath := fmt.Sprintf("c%d", atomID/1000)
	htDocs := os.Getenv("CONTENT_HTDOCS")
	atomDir := filepath.Join(htDocs, "content", subpath, atomIDStr)
	if ok, _ := pathExists(atomDir); !ok { // dir does not exist
		if err := os.Mkdir(atomDir, 0755); err != nil {
			return "", fmt.Errorf("Can't mkdir for atom %d", atomID)
		}
	}
	return atomDir, nil
}

//FixAtomPems - will mirror atom files
func FixAtomPems(job worker.Job) ([]byte, error) {
	data := string(job.Data())
	atomID, err := strconv.Atoi(data)
	if err != nil {
		return []byte("Invalid atom id:" + err.Error()), nil
	}
	atomDir, err := computeAtomPath(atomID)
	//log.Println("Got atom", atomID)
	//log.Println("  ", atomDir)
	if err != nil {
		return []byte("Can't get atom dir: " + err.Error()), nil
	}

	out := ""
	// lets do the walk
	err = filepath.Walk(atomDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			out += fmt.Sprintf("unable to fix [%s]\n", path)
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			//fmt.Printf("visited dir: %q\n", path)
			if err := os.Chmod(path, 0755); err != nil {
				out += fmt.Sprintf("  unable to dir fix [%s]\n", path)
			} else {
				out += fmt.Sprintf("  fixed dir [%s]\n", path)
			}
		} else if info.Mode().IsRegular() {
			//fmt.Printf("visited regular file: %q\n", path)
			//fmt.Printf("  pems: %q\n", info.Mode().Perm())
			if err := os.Chmod(path, 0644); err != nil {
				log.Printf("chmod(%s) = %+v\n", path, err.Error())
				out += fmt.Sprintf("  unable to fix file [%s]\n", path)
			} else {
				out += fmt.Sprintf("  fixed file [%s]\n", path)
			}

		} else {
			fmt.Printf("visited other type of file: %q\n", path)
			//fmt.Printf("  pems: %q\n", info.Mode().Perm())
			out += fmt.Sprintf("  not fixing [%s]\n", path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", atomDir, err)
		out += fmt.Sprintf(" error fixing in [%s]: %s\n", atomDir, err.Error())
	}

	return []byte(out), nil
}

func initDB() (*sql.DB, error) {
	var dbh *sql.DB
	var err error
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = os.Getenv("USER")
	}

	dbPass := os.Getenv("DB_PASS")

	dbName := os.Getenv("DB_DATABASE")
	if "" == dbName {
		return dbh, fmt.Errorf("DB_DATABASE not set")
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	// default collation is 'utf8mb4_general_ci'
	dbh, err = sql.Open("mysql", dbUser+":"+dbPass+"@tcp("+dbHost+")/"+dbName)
	if err != nil {
		log.Println(dbUser + ":XXX" + "@tcp(" + dbHost + ")/" + dbName)
		log.Println(err.Error())
		return dbh, err
	}

	// Open doesn't open a connection, so let's Ping() our db
	err = dbh.Ping()
	if err != nil {
		log.Println(dbUser + ":XXX" + "@tcp(" + dbHost + ")/" + dbName)
		//panic(err.Error()) // proper error handling instead of panic in your app
		log.Println(err.Error())
		return dbh, err
	}

	return dbh, nil
}

//SynchAtomFiles - will mirror atom files
func SynchAtomFiles(job worker.Job) ([]byte, error) {
	data := string(job.Data())
	log.Println("Got ", data)
	atomID, err := strconv.Atoi(data)
	if err != nil {
		return []byte("Invalid atom id:" + err.Error()), nil
	}
	atomDir, err := computeAtomPath(atomID)
	//log.Println("Got atom", atomID)
	log.Println("  we'll store data in ", atomDir)
	if err != nil {
		return []byte("Can't get atom dir: " + err.Error()), nil
	}

	dbh, err := initDB()
	//log.Println("DB init'ed")
	if err != nil {
		return []byte(err.Error()), nil
	}
	defer dbh.Close()

	atomGetter := Atoms{DB: dbh}
	atom, err := atomGetter.GetByID(atomID)
	if err != nil {
		panic(err.Error())
	}
	//fmt.Printf("atom = %+v\n", atom)
	log.Printf("[%s] %s\n", *atom.ID, *atom.Name)

	out := ""
	for _, ad := range atom.Downloads {
		if *ad.Type == "PDF" {
			//log.Println(" Type: ", *ad.Type)
			//log.Println(" Label:", *ad.Label)
			log.Println(" Path: ", *ad.Path)
			//log.Println("")

			url := "https://dnalc.cshl.edu" + *ad.Path
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("err.Error() = %+v\n", err.Error())
				log.Printf("Status: %s", resp.Status)
				continue
			}
			defer resp.Body.Close()

			fullAtomPath := filepath.Join(atomDir, path.Base(*ad.Path))
			out += fullAtomPath + "\n"

			// Create the file
			fh, err := os.Create(fullAtomPath)
			if err != nil {
				out += "Error: " + err.Error() + "\n"
			}
			defer fh.Close()

			// Write the body to file
			_, err = io.Copy(fh, resp.Body)
			if err != nil {
				out += "Error: " + err.Error() + "\n"
			} else {
				out += "Saved file to " + fullAtomPath
			}
		}
	}

	//log.Println("about to return: " + out)
	return []byte(out), nil
}
