package fsutil

type Archive interface {
	GetFileNames(onlyImages bool) ([]string, error)
	GetFirstFileName(onlyImages bool) (string, error)
	ImagesToMap() (map[int]string, error)
	ReadFile(name string) ([]byte, error)
	Close() error
}

func OpenArchive(path string) (Archive, error) {
	// Check file extension or magic bytes
	if Is7z(path) {
		return NewSevenZipArchive(path)
	}
	return NewZipArchive(path)
}
