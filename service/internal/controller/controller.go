package controller

import (
	"fmt"
	"io"
	"log"
	"mailcode/service/internal/mailwatcher"
)

type WatcherCtl struct{}

func (*WatcherCtl) ListEmails(repo *mailwatcher.Repository) int {
	mbs, err := repo.GetAllMailboxes()
	if err != nil {
		log.Println(err)
		return 1
	}

	for mb := mbs.Front(); mb != nil; mb = mb.Next() {
		fmt.Println(mb.Value.(mailwatcher.Mailbox).ToString())
	}

	return 0
}

func (*WatcherCtl) AddEmail(repo *mailwatcher.Repository, mb *mailwatcher.Mailbox) int {
	err := repo.AddMailbox(mb)
	if err != nil {
		log.Println(err)
		return 1
	}
	return 0
}

func (*WatcherCtl) RemoveEmail(repo *mailwatcher.Repository, email string) int {
	err := repo.RemoveMailbox(email)
	if err != nil {
		log.Println(err)
		return 1
	}
	return 0
}

func GetMsg(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
		println("Client got:", string(buf[0:n]))
	}

	println("Done listening")
}

func SendMsg(w io.Writer, msg *mailwatcher.Message) {
	m, err := mailwatcher.Serialize(msg)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = w.Write(m)
	if err != nil {
		log.Fatalln(err)
	}
}
