package entity

type FileData struct {
	AbsPath string
	Size    int64
	Hash    string
	Handled bool
}
