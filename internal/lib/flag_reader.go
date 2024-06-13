package lib

import "flag"

// default flags values if they are not set during launch
const (
	defaultRootPath     = "."
	defaultWorkersLimit = 1
	defaultConnString   = "files.db"
	defaultShowResult   = false
	defaultOwerwrite    = false
)

// get and return flags from launch arguments
func GetFlagValues() (rootPath string, workersLimit int, connString string, showResult bool, overwrite bool) {

	// path to start directory
	flag.StringVar(&rootPath, "path", defaultRootPath, "path to start directory")

	// limit of goroutines calculating checksum
	flag.IntVar(&workersLimit, "workers", defaultWorkersLimit, "limit of checksum calculating goroutines")

	// connection string to database
	flag.StringVar(&connString, "db", defaultConnString, "connection string to database")

	// wheter show found duplicates
	flag.BoolVar(&showResult, "show", defaultShowResult, "show result (duplicates names and count)")

	// wheter show found duplicates
	flag.BoolVar(&overwrite, "overwrite", defaultOwerwrite, "overwrite file than contains files data")

	// Parse the flags
	flag.Parse()

	return
}
