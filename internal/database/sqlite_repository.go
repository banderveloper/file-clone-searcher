package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/banderveloper/goFileCloneSearcher/internal/entity"
)

type SqliteRepository struct {
	db    *sql.DB
	files []*entity.FileData
}

func NewSqliteRepository(db *sql.DB) Repository {
	return &SqliteRepository{db: db}
}

func (r *SqliteRepository) EnsureTableCreated() error {

	query := `CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		size INTEGER,
		hash TEXT);`

	_, err := r.db.Exec(query)

	return err
}

func (r *SqliteRepository) AddFileData(fd *entity.FileData) {
	r.files = append(r.files, fd)
}

func (r *SqliteRepository) Commit() error {

	if len(r.files) == 0 {
		return fmt.Errorf("commit error, nothing to commit")
	}

	var buffer bytes.Buffer
	buffer.WriteString("INSERT INTO files(name, size, hash) VALUES")

	for _, fd := range r.files {

		pathParts := strings.Split(fd.AbsPath, string(os.PathSeparator))
		fileName := pathParts[len(pathParts)-1]

		buffer.WriteString(fmt.Sprintf("('%s', %d, '%s'), ", fileName, fd.Size, fd.Hash))
	}

	query := buffer.String()
	query = strings.TrimRight(query, ", ")

	_, err := r.db.Exec(query)

	return err
}
