package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/banderveloper/file-clone-searcher/internal/entity"
)

// child of Repository struct
type SqliteRepository struct {
	db    *sql.DB
	files []*entity.FileData
}

func NewSqliteRepository(db *sql.DB) Repository {
	return &SqliteRepository{db: db}
}

// create table if not exists
func (r *SqliteRepository) EnsureTableCreated(overwrite bool) error {

	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		size INTEGER,
		hash TEXT);`)

	if err != nil {
		return err
	}

	// if --overwrite flag is not set, clear existing table
	if overwrite {

		_, err := r.db.Exec("DELETE FROM files")
		if err != nil {
			return err
		}
	}

	return nil
}

// add file data to inmemory store
func (r *SqliteRepository) AddFileData(fd *entity.FileData) {
	r.files = append(r.files, fd)
}

// send all accumulated files data to db
func (r *SqliteRepository) Commit() error {

	if len(r.files) == 0 {
		return fmt.Errorf("commit error, nothing to commit")
	}

	var buffer bytes.Buffer
	buffer.WriteString("INSERT INTO files(name, size, hash) VALUES")

	for _, fd := range r.files {

		// fd has absolute path, take only last part from path with filename and extension
		pathParts := strings.Split(fd.AbsPath, string(os.PathSeparator))
		fileName := pathParts[len(pathParts)-1]

		buffer.WriteString(fmt.Sprintf("('%s', %d, '%s'), ", fileName, fd.Size, fd.Hash))
	}

	// trim last comma and space
	query := buffer.String()
	query = strings.TrimRight(query, ", ")

	_, err := r.db.Exec(query)

	return err
}

// Discover and get duplicated files from commited data
func (r *SqliteRepository) GetDuplicates() ([]*entity.CloneData, error) {

	duplicates := make([]*entity.CloneData, 0)

	// query for get duplicates of files by name, size and hash
	query := `SELECT name, size, hash, count(*) AS count 
		FROM files
		GROUP BY name, size, hash
		HAVING count > 1
		ORDER BY count DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {

		var dupl = &entity.CloneData{}
		rows.Scan(&dupl.Name, &dupl.Size, &dupl.Hash, &dupl.Count)
		duplicates = append(duplicates, dupl)
	}

	return duplicates, nil
}
