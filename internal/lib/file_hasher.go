package lib

import (
	"encoding/hex"
	"hash/adler32"
	"io"
	"os"

	"github.com/banderveloper/fileCloneSearcher/internal/entity"
)

// hashing algo
// change adler32 to another to change hashing algorythm
var hasher = adler32.New()

// fill file checksum for given file
// after calculating - send done fileData to channel
// function needs to file fd.Hash and fd.Handled fields
func SetCheckSum(fd *entity.FileData, quotaCh chan struct{}, fileEntitiesCh chan<- *entity.FileData) {

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

	if _, err := io.Copy(hasher, file); err != nil {
		//log.Printf("Error copying file content of %s\n", fe.absPath)
		fileEntitiesCh <- fd
		<-quotaCh
		return
	}

	// dont calculate checksum of empty file
	if fd.Size > 0 {
		checksum := hasher.Sum(nil)
		fd.Hash = hex.EncodeToString(checksum)
	}

	fd.Handled = true
	fileEntitiesCh <- fd
	<-quotaCh
}
