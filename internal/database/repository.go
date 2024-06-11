package database

import "github.com/banderveloper/goFileCloneSearcher/internal/entity"

type Repository interface {
	EnsureTableCreated() error
	AddFileData(fd *entity.FileData)
	Commit() error
}
