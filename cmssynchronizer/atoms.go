package cmssynchronizer

import (
	"database/sql"
	"fmt"
	"strings"
)

// Atom type for the atoms in DNALC's CMS
type Atom struct {
	ID             *string        `json:"id"`
	Name           *string        `json:"title"`
	ShortDesc      *string        `json:"description"`
	Type           *string        `json:"type"`
	Permalink      *string        `json:"permalink"`
	Narative       *string        `json:"transcript"`
	Keywords       *string        `json:"keywords"`
	Tags           *string        `json:"tags"`
	MediaText      *string        `json:"media_text"`
	Thumbnail      *string        `json:"featured_image"`
	AnimationHTML5 *string        `json:"animation_html5"`
	Downloads      []AtomDownload `json:"downloads"`
	Public         *string        `json:"public"`
}

type AtomDownload struct {
	Type  *string `json:"type"`
	Label *string `json:"path"`
	Path  *string `json:"path"`
}

type Atoms struct {
	DB *sql.DB
}

const sqlSingleAtomQueryTmpl = `SELECT at_name, at_type, at_public FROM atoms WHERE at_id = ?`
const sqlAtomDownloadsQueryTmpl = `SELECT path, label, type FROM atom_downloads where atom_id = ?`

func getAtomDownloads(dbh *sql.DB, id int32) []AtomDownload {

	sth, err := dbh.Query(sqlAtomDownloadsQueryTmpl, id)
	if err != nil {
		panic(err.Error())
	}

	downloads := []AtomDownload{}
	for sth.Next() {
		var ad AtomDownload
		//var dPath, dLabel, dType string
		err := sth.Scan(&ad.Path, &ad.Label, &ad.Type)
		//err := sth.Scan(&dPath, &dLabel, &dType)
		//fmt.Printf("ad = {%s, %s, %s}\n", dPath, dLabel, dType)
		if err != nil {
			return downloads
		}
		//downloads = append(downloads, AtomDownload{&dType, &dLabel, &dPath})
		downloads = append(downloads, ad)
	}
	return downloads
}

func (a *Atoms) GetByID(atomID int32) (Atom, error) {

	dbh := a.DB

	var atom Atom
	row := dbh.QueryRow(sqlSingleAtomQueryTmpl, atomID)

	var atomIDStr string
	err := row.Scan(&atom.Name, &atom.Type, &atom.Public)
	if err != nil {
		panic(err.Error())
	}

	atomIDStr = fmt.Sprintf("%d", atomID)
	atom.ID = &atomIDStr
	*atom.Name = strings.TrimSpace(*atom.Name)

	//atomDownloads := getAtomDownloads(dbh, atomID)
	//fmt.Printf("downloads for [%s]: %v\n", atomID, atomDownloads)
	atom.Downloads = getAtomDownloads(dbh, atomID)

	return atom, nil
}
