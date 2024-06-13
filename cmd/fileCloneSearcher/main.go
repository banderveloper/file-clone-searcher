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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/banderveloper/fileCloneSearcher/internal/database"
	"github.com/banderveloper/fileCloneSearcher/internal/entity"
	"github.com/banderveloper/fileCloneSearcher/internal/lib"

	_ "modernc.org/sqlite"
)

// Recursively get folder's nested items - files and folders
// path - folder path to analyze
// quotaCh - channel for limiting one-time checksum goroutines count
// filesDataCh - channel for inserting files with calculated checksum (chan is passed to checksum method)
func analyzeDir(path string, quotaCh chan struct{}, filesDataCh chan<- *entity.FileData) {

	// get info about current folder
	curDir, err := os.Open(path)
	if err != nil {
		return
	}

	defer curDir.Close()

	// folded files and folders
	foldedItems, err := curDir.Readdir(0)
	if err != nil {
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
	rootPath, workersLimit, connString, showResult := lib.GetFlagValues()

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
		for {
			select {

			// if it is file data in channel - add it to inmemory
			case fd := <-filesDataCh:
				if fd.Handled {
					repos.AddFileData(fd)
				}

			// if chan is empty - wait 10 nanoseconds (prevents deadlock)
			case <-time.After(time.Nanosecond * 10):

			}
		}
	}()

	// start recursively analyzing file system
	analyzeDir(rootPath, quotaCh, filesDataCh)

	// close quota channel after scanning all folders
	close(quotaCh)

	// save accumulated files info to db
	err = repos.Commit()
	if err != nil {
		panic(err)
	}

	// if --show flag is set, show duplicates files in terminal
	if showResult {

		duplicates, err := repos.GetDuplicates()
		if err != nil {
			panic(err)
		}

		for _, dupl := range duplicates {
			fmt.Println(dupl.ToString())
		}
	}

	log.Println("Done!")
}
