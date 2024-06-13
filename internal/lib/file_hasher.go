package lib

import (
	"encoding/hex"
	"hash/adler32"
	"io"
	"os"

	"github.com/banderveloper/fileCloneSearcher/internal/entity"
)

// fill file checksum for given file
// after calculating - send done fileData to channel
// function needs to file fd.Hash and fd.Handled fields
func SetCheckSum(fd *entity.FileData, quotaCh chan struct{}, fileEntitiesCh chan<- *entity.FileData) {

	// open current file
	// if error - send file with Handled:false and empty checksum
	// and release quota
	file, err := os.Open(fd.AbsPath)
	if err != nil {
		<-quotaCh
		return
	}

	defer file.Close()

	// hashing algo
	// change adler32 to another to change hashing algorythm
	var hasher = adler32.New()

	if _, err := io.Copy(hasher, file); err != nil {
		<-quotaCh
		return
	}

	// calculate checksum only for non-empty files
	if fd.Size > 0 {
		checksum := hasher.Sum(nil)
		fd.Hash = hex.EncodeToString(checksum)
	} else {
		<-quotaCh
		return
	}

	fd.Handled = true
	fileEntitiesCh <- fd
	<-quotaCh
}
