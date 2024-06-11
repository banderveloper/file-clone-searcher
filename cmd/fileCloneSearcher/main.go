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
	"hash/adler32"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/banderveloper/goFileCloneSearcher/internal/database"
	"github.com/banderveloper/goFileCloneSearcher/internal/entity"

	_ "modernc.org/sqlite"
)

func analyzeDir(path string, quotaCh chan struct{}, filesDataCh chan<- *entity.FileData) {

	curDir, err := os.Open(path)
	if err != nil {
		//log.Printf("Error during opening directory %s\n", path)
		return
	}

	defer curDir.Close()

	foldedItems, err := curDir.Readdir(0)
	if err != nil {
		//log.Printf("Error during getting folded items in directory %s\n", foldedItems)
		return
	}

	for _, folded := range foldedItems {

		curDir.Chdir()

		if folded.IsDir() {
			analyzeDir(folded.Name(), quotaCh, filesDataCh)
			continue
		}

		absPath, err := filepath.Abs(folded.Name())
		if err != nil {
			//log.Printf("Error during getting abs path of file %s/%s\n", path, folded.Name())
			continue
		}

		entity := &entity.FileData{
			AbsPath: absPath,
			Size:    folded.Size(),
		}

		quotaCh <- struct{}{}
		go setControlSum(entity, quotaCh, filesDataCh)
	}
}

func setControlSum(fd *entity.FileData, quotaCh chan struct{}, fileEntitiesCh chan<- *entity.FileData) {

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

	if fd.Size > 0 {
		checksum := hash.Sum(nil)
		fd.Hash = hex.EncodeToString(checksum)
	}

	fd.Handled = true
	fileEntitiesCh <- fd
	<-quotaCh
}

func main() {

	db, err := sql.Open("sqlite", "files.db")
	if err != nil {
		panic(err)
	}
	repos := database.NewSqliteRepository(db)
	repos.EnsureTableCreated()

	workersLimit := 3
	rootPath := "/home/nikita/Documents"

	filesDataCh := make(chan *entity.FileData, workersLimit)
	quotaCh := make(chan struct{}, workersLimit)

	go func() {
		for fd := range filesDataCh {
			repos.AddFileData(fd)
		}
	}()

	analyzeDir(rootPath, quotaCh, filesDataCh)

	err = repos.Commit()
	if err != nil {
		panic(err)
	}

	log.Println("Done!")
}
