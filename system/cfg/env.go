package cfg

import (
	"log"
	"os"
	"path/filepath"
)

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't find working directory", err)
	}
	return wd
}

func DataDir() string {
	dataDir := os.Getenv("KUDZU_DATA_DIR")
	if dataDir == "" {
		return getWd()
	}
	return dataDir
}

func TlsDir() string {
	tlsDir := os.Getenv("KUDZU_TLS_DIR")
	if tlsDir == "" {
		tlsDir = filepath.Join(getWd(), "system", "tls")
	}
	return tlsDir
}

func AdminStaticDir() string {
	staticDir := os.Getenv("KUDZU_ADMINSTATIC_DIR")
	if staticDir == "" {

		staticDir = filepath.Join(getWd(), "system", "admin", "static")
	}
	return staticDir
}

func UploadDir() string {
	uploadDir := os.Getenv("KUDZU_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = filepath.Join(DataDir(), "uploads")
	}
	return uploadDir
}

func SearchDir() string {
	searchDir := os.Getenv("KUDZU_SEARCH_DIR")
	if searchDir == "" {
		searchDir = filepath.Join(DataDir(), "search")
	}
	return searchDir
}
