// Напиши искалку дубликатов файлов в системе. По имени, размеру и контрольной сумме.
// сделай так, чтоб все вычислялось «параллельно», бери по N файлов, пихай в горутины, ответ от них через каналы собирай, складывай в бд например.
// Потом уже придумать запрос, который найдет тебе дубли в бд
// В бд клади все файлы, для каждого:
// Полный путь, имя файла, размер, контрольная сумма
// Потом запросом с группировкой легко сможешь найти дубли по контрольной сумме, размеру и имени
// Под результат запусти отдельную горутину, пусть батчи в бд собирает

package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/banderveloper/fileCloneSearcher/internal/database"
	"github.com/banderveloper/fileCloneSearcher/internal/entity"
	"github.com/banderveloper/fileCloneSearcher/internal/lib"

	_ "modernc.org/sqlite"
)

// Recursively get folder's nested items - files and folders
func analyzeDir(path string, quotaCh chan struct{}, filesDataCh chan<- *entity.FileData) {

	// get info about current folder
	curDir, err := os.Open(path)
	if err != nil {
		//log.Printf("Error during opening directory %s\n", path)
		return
	}

	defer curDir.Close()

	// folded files and folders
	foldedItems, err := curDir.Readdir(0)
	if err != nil {
		//log.Printf("Error during getting folded items in directory %s\n", foldedItems)
		return
	}

	for _, folded := range foldedItems {

		curDir.Chdir()

		// if its folded directory - recursively deep into it
		if folded.IsDir() {
			analyzeDir(folded.Name(), quotaCh, filesDataCh)
			continue
		}

		// if its file - get absolute path
		absPath, err := filepath.Abs(folded.Name())
		if err != nil {
			//log.Printf("Error during getting abs path of file %s/%s\n", path, folded.Name())
			continue
		}

		// fill some data about this file, and if quota is available, start goroutine for calculating checksum
		entity := &entity.FileData{
			AbsPath: absPath,
			Size:    folded.Size(),
		}

		quotaCh <- struct{}{}
		go lib.SetCheckSum(entity, quotaCh, filesDataCh)
	}
}

func main() {

	// Get run flags
	rootPath, workersLimit, connString, _ := lib.GetFlagValues()

	// channel with filled files data
	filesDataCh := make(chan *entity.FileData, workersLimit)
	// quota channel for limiting checksum goroutines
	quotaCh := make(chan struct{}, workersLimit)

	// connect and initialize db
	db, err := sql.Open("sqlite", connString)
	if err != nil {
		panic(err)
	}
	repos := database.NewSqliteRepository(db)
	repos.EnsureTableCreated()

	// goroutine for reading channel with done files data and saving it to inmemory
	go func() {
		for fd := range filesDataCh {
			if fd.Handled {
				repos.AddFileData(fd)
			}
		}
	}()

	// start recursively analyzing file system
	analyzeDir(rootPath, quotaCh, filesDataCh)

	close(quotaCh)

	// save accumulated files info to db
	err = repos.Commit()
	if err != nil {
		panic(err)
	}

	log.Println("Done!")
}
