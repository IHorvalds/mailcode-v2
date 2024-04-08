package main

import (
	"flag"
	"log"
	"mailcode/service/internal/mailwatcher"
	"mailcode/service/internal/watcher"
	"os"
	"strings"
)

func main() {
	var databasePathFlag = flag.String("db", "", "Path to the emails database. Defaults to the directory this utility is run from")
	var configFileFlag = flag.String("config", mailwatcher.DefaultConfigFile(), "Configuration file for subjects to search for and regexes to extract auth codes")

	flag.Parse()

	confPath := *configFileFlag
	conf, err := mailwatcher.LoadConfig(confPath)
	if err != nil {
		log.Fatalln(err)
	}

	dbPath := conf.DatabasePath

	// Database path given at CLI has priority
	if strings.Compare(*databasePathFlag, "") != 0 {
		dbPath = *databasePathFlag
	}

	if strings.Compare(dbPath, "") == 0 {
		log.Fatalln("Path to the database must be specified either in the cli arguments or in the config file")
	}

	repo, err := mailwatcher.OpenRepository(dbPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer repo.Close()

	code := new(watcher.Watcher).Run(&repo, &conf)
	os.Exit(code)
}
