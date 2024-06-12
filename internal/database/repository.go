package database

import "github.com/banderveloper/fileCloneSearcher/internal/entity"

type Repository interface {
	// create table if not exists
	EnsureTableCreated() error

	// add file data to inmemory store
	AddFileData(fd *entity.FileData)

	// send all accumulated files data to db
	Commit() error
}
