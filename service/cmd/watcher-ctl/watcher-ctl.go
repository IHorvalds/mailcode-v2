package main

import (
	"flag"
	"log"
	"mailcode/service/internal/controller"
	"mailcode/service/internal/mailwatcher"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	var databasePathFlag = flag.String("db", "", "Path to the emails database. Defaults to the directory this utility is run from")
	var configFileFlag = flag.String("config", mailwatcher.DefaultConfigFile(), "Configuration file for subjects to search for and regexes to extract auth codes")

	// Commands
	var listFlag = flag.Bool("list", false, "List the current emails")
	var addFlag = flag.Bool("add", false, "Add a new email to the list")
	var deleteFlag = flag.Bool("delete", false, "Remove email from the list")

	var sendMsgFlag = flag.String("msg", "", "Message to send")

	// Mailbox info
	var emailFlag = flag.String("email", "", "")
	var passwordFlag = flag.String("password", "", "")
	var serverFlag = flag.String("server", "", "")
	var useTLSFlag = flag.Bool("with-tls", true, "")
	var portFlag = flag.Int("port", 0, "")

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

	ctl := new(controller.WatcherCtl)
	if *listFlag {

		os.Exit(ctl.ListEmails(&repo))
	}

	if *addFlag {
		mb := mailwatcher.Mailbox{
			Email:    *emailFlag,
			Password: *passwordFlag,
			Server:   *serverFlag,
			Port:     int32(*portFlag),
			UseSSL:   *useTLSFlag,
		}

		os.Exit(ctl.AddEmail(&repo, &mb))
	}

	if *deleteFlag {
		os.Exit(ctl.RemoveEmail(&repo, *emailFlag))
	}

	// Just for test
	c, err := net.Dial("unix", "/tmp/mailwatcher.sock")
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	go controller.GetMsg(c)
	var cmd mailwatcher.Action
	switch *sendMsgFlag {
	case "Watch":
		cmd = mailwatcher.Watch
	case "WatchAll":
		cmd = mailwatcher.WatchAll
	case "Stop":
		cmd = mailwatcher.Stop
	case "StopAll":
		cmd = mailwatcher.StopAll
	default:
		cmd = mailwatcher.ConnectionError
	}

	msg := mailwatcher.Message{
		Cmd: cmd,
		Params: map[string]interface{}{
			"email": *emailFlag,
		},
	}
	controller.SendMsg(c, &msg)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Fatalln("One of list, add or remove must be specified")
}
