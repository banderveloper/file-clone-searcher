package database

import "github.com/banderveloper/file-clone-searcher/internal/entity"

type Repository interface {
	// create table if not exists
	EnsureTableCreated(overwrite bool) error

	// add file data to inmemory store
	AddFileData(fd *entity.FileData)

	// send all accumulated files data to db
	Commit() error

	// Discover and get duplicated files from commited data
	GetDuplicates() ([]*entity.CloneData, error)
}
