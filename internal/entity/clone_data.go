package entity

import "fmt"

// duplicated file info
type CloneData struct {
	Name  string // name of the duplicated file
	Size  int64
	Hash  string // checksum
	Count int    // count of duplicates
}

func (cd *CloneData) ToString() string {
	return fmt.Sprintf("%s (%db) | Hash: %s | Count: %d", cd.Name, cd.Size, cd.Hash, cd.Count)
}
