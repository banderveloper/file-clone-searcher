package lib

import "flag"

const (
	defaultRootPath     = "."
	defaultWorkersLimit = 1
	defaultConnString   = "files.db"
	defaultShowResult   = false
)

func GetFlagValues() (rootPath string, workersLimit int, connString string, showResult bool) {

	// path to start directory
	flag.StringVar(&rootPath, "path", defaultRootPath, "Path to start directory")

	// limit of goroutines calculating checksum
	flag.IntVar(&workersLimit, "workers", defaultWorkersLimit, "Limit of checksum calculating goroutines")

	// connection string to database
	flag.StringVar(&connString, "db", defaultConnString, "Connection string to database")

	// wheter show found duplicates
	flag.BoolVar(&showResult, "show", defaultShowResult, "Show result (duplicates names and count)")

	// Parse the flags
	flag.Parse()

	return
}
