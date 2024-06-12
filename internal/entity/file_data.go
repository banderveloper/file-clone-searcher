package entity

type FileData struct {
	AbsPath string
	Size    int64
	Hash    string // checksum
	Handled bool   // is file opened and calculated checksum
}
