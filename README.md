# file-clone-searcher

This golang utility recursively searches duplicating files in the given folder, then optionally shows it to terminal.

## Solution
Application recursively search all files in given folder, and starts quota-limited goroutines for calculating file checksum. When checksum is ready - file name, size and checksum writes to database. After collecting all files, optionally application finds duplicating files by sql query and shows them to the terminal

## What's used
- Quota channel to limiting concurrently running checksum calculating goroutines
- Polymorphic repository-pattern system to interacting with the database.
- Recursive file system search
- File content handling and hashing

## How to run

### Non-compiled .go file
- Ensure having installed [go compiler](https://go.dev/dl/)
- Clone repository ```git clone https://github.com/banderveloper/file-clone-searcher```
- Execute ```go run cmd/searcher/main.go [FLAGS]```

### Compiled binary
- Go to [repository releases](https://github.com/banderveloper/file-clone-searcher/releases)
- Download last release
- Execute ```./searcher_linuxamd64 [FLAGS]```  or  ```.\searcher_windowsamd64.exe [FLAGS]```

### Flags
Utility accepts launch flags. Get information about available flags using --help launch flag 

## Contact
If you have any questions - create issue, or contact me using [Instagram](https://www.instagram.com/banderveloper) or Telegram (@bandernik)
