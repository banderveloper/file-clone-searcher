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
	"encoding/hex"
	"flag"
	"hash/adler32"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/banderveloper/fileCloneSearcher/internal/database"
	"github.com/banderveloper/fileCloneSearcher/internal/entity"

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
		go setCheckSum(entity, quotaCh, filesDataCh)
	}
}

// fill file checksum for given file
// after calculating - send done fileData to channel
// function needs to file fd.Hash and fd.Handled fields
func setCheckSum(fd *entity.FileData, quotaCh chan struct{}, fileEntitiesCh chan<- *entity.FileData) {

	// open current file
	// if error - send file with Handled:false and empty checksum
	// and release quota
	file, err := os.Open(fd.AbsPath)
	if err != nil {
		//log.Printf("Error during opening file %s\n", fe.absPath)
		fileEntitiesCh <- fd
		<-quotaCh
		return
	}

	defer file.Close()

	hash := adler32.New()

	if _, err := io.Copy(hash, file); err != nil {
		//log.Printf("Error copying file content of %s\n", fe.absPath)
		fileEntitiesCh <- fd
		<-quotaCh
		return
	}

	// dont calculate checksum of empty file
	if fd.Size > 0 {
		checksum := hash.Sum(nil)
		fd.Hash = hex.EncodeToString(checksum)
	}

	fd.Handled = true
	fileEntitiesCh <- fd
	<-quotaCh
}

func main() {

	// Define flags

	// start path
	var rootPath string
	flag.StringVar(&rootPath, "path", ".", "Path to start directory")

	// limit of goroutines calculating checksum
	var workersLimit int
	flag.IntVar(&workersLimit, "workers", 1, "Limit of checksum calculating goroutines")

	// connection string to database
	var connString string
	flag.StringVar(&connString, "db", "files.db", "Connection string to database")

	// wheter show found duplicates
	var showResult bool
	flag.BoolVar(&showResult, "show", false, "Show result (duplicates names and count)")

	// Parse the flags
	flag.Parse()

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

	// save accumulated files info to db
	err = repos.Commit()
	if err != nil {
		panic(err)
	}

	log.Println("Done!")
}
