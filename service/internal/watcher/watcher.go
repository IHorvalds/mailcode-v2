package watcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mailcode/service/internal/mailwatcher"
	"net"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

type Watcher struct {
	ctxs        *map[string]*mailwatcher.MailboxContext
	repo        *mailwatcher.Repository
	config      *mailwatcher.Configuration
	codeChannel chan mailwatcher.EmailCode
}

// Watcher methods
func (w *Watcher) Run(repo *mailwatcher.Repository, config *mailwatcher.Configuration) int {

	codeChannel := make(chan mailwatcher.EmailCode)
	defer close(codeChannel)
	mbs, err := repo.GetAllMailboxes()
	if err != nil {
		log.Print(err)
		return 1
	}

	mailboxes, err := mailwatcher.WatchMailboxes(mbs, config, codeChannel)
	if err != nil {
		log.Print(err)
		return 1
	}

	w.ctxs = &mailboxes
	w.repo = repo
	w.config = config
	w.codeChannel = codeChannel

	// Open UNIX socket, accept connections
	// Keep connections alive
	// Accept messages, handle based on cmd
	//
	// When sending a message (Code or Error), broadcast to all connections
	s := NewServer("/tmp/mailwatcher.sock")
	s.watcher = w
	go s.Serve()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for code := range codeChannel {
			log.Printf("Code is %s\n", code)
			msg := mailwatcher.Message{
				Cmd: mailwatcher.Code,
				Params: map[string]interface{}{
					"code":   code.Code,
					"sender": code.Sender,
				},
			}

			msgBytes, err := mailwatcher.Serialize(&msg)

			if err != nil {
				log.Println(err)
			}

			s.mux.Lock()
			for el := s.connections.Front(); el != nil; el = el.Next() {
				conn, ok := el.Value.(net.Conn)
				if !ok {
					log.Panicf("Unexpected type %s in connections list\n", reflect.TypeOf(el.Value))
				}
				_, err := conn.Write(msgBytes)
				if err != nil {
					log.Println("Failed to send a code to a connection. Removing connection...")
					s.connections.Remove(el)
				}
			}
			s.mux.Unlock()
		}
	}()

	<-c
	log.Println("\nShutting down...")
	s.Stop()
	mailwatcher.StopWatchingMailboxes(s.watcher.ctxs)
	stopChan := make(chan struct{})

	go func() {
		for {
			bStopped := true
			for _, ctx := range *s.watcher.ctxs {
				if mailwatcher.IsRunning(ctx) {
					bStopped = false
				}
			}

			if bStopped {
				break
			}
		}

		stopChan <- struct{}{}
	}()

	select {
	case <-time.After(time.Duration(10 * time.Second)):
		log.Println("Timeout waiting to stop watching mailboxes")
	case <-stopChan:
		break
	}
	return 0
}

func (w *Watcher) handleMessage(msg *mailwatcher.Message) (*mailwatcher.Message, error) {
	action, err := msg.Cmd.ToString()
	if err != nil {
		return nil, fmt.Errorf("invalid action %d", msg.Cmd)
	}

	log.Printf("Handling message action %s\n", action)
	switch msg.Cmd {
	case mailwatcher.Add:
		// Add email
		mb, err := map2Mailbox(&msg.Params)
		if err != nil {
			return nil, errors.New("failed to parse mailbox from message params")
		}
		w.repo.AddMailbox(mb)
	case mailwatcher.Remove:
		// Remove email
		em, ok := msg.Params["email"].(string)
		if !ok {
			return nil, errors.New("failed to parse email from message params")
		}
		w.repo.RemoveMailbox(em)
	case mailwatcher.Watch:
		// Watch email
		em, ok := msg.Params["email"].(string)
		if !ok {
			return nil, errors.New("failed to parse email from message params")
		}
		mb, err := w.repo.GetMailbox(em)
		if err != nil {
			return nil, err
		}
		ctx, exists := (*w.ctxs)[em]
		if !exists || !mailwatcher.IsRunning(ctx) {
			(*w.ctxs)[em] = mailwatcher.WatchMailbox(&mb, w.config, w.codeChannel)
		}
	case mailwatcher.WatchAll:
		// Start watching all emails
		//
		// Get all emails from repo
		// whichever ones aren't in w.ctxs,
		// add them => go watch() them
		mbs, err := w.repo.GetAllMailboxes()
		if err != nil {
			return nil, err
		}
		for el := mbs.Front(); el != nil; el = el.Next() {
			mb := el.Value.(*mailwatcher.Mailbox)
			ctx, exists := (*w.ctxs)[mb.Email]
			if !exists || !mailwatcher.IsRunning(ctx) {
				(*w.ctxs)[mb.Email] = mailwatcher.WatchMailbox(mb, w.config, w.codeChannel)
			}
		}
	case mailwatcher.Stop:
		// Stop watching email
		em, ok := msg.Params["email"].(string)
		if !ok {
			return nil, errors.New("failed to parse email from message params")
		}
		ctx, exists := (*w.ctxs)[em]
		if exists {
			mailwatcher.StopWatchingMailbox(ctx)
			delete((*w.ctxs), em)
		}
	case mailwatcher.StopAll:
		// Stop watching all emails. Don't exit
		//
		// Signal all the ctxs
		// clear out the ctxs
		mailwatcher.StopWatchingMailboxes(w.ctxs)
	case mailwatcher.GetMailbox:
		em, ok := msg.Params["email"].(string)
		if !ok {
			return nil, errors.New("failed to parse email from message params")
		}

		mb, err := w.repo.GetMailbox(em)
		if err != nil {
			e := fmt.Sprintf("email '%s' not found", em)
			return &mailwatcher.Message{
					Cmd: mailwatcher.GetMailbox,
					Params: map[string]interface{}{
						"error": e,
					},
				},
				errors.New(e)
		}

		return &mailwatcher.Message{
			Cmd:    mailwatcher.GetMailbox,
			Params: mailbox2map(&mb),
		}, nil
	case mailwatcher.GetAllMailboxes:
		mbs, err := w.repo.GetAllMailboxes()
		if err != nil {
			e := "error getting emails"
			return &mailwatcher.Message{
				Cmd: mailwatcher.GetAllMailboxes,
				Params: map[string]interface{}{
					"error": e,
				},
			}, errors.New(e)
		}

		return &mailwatcher.Message{
			Cmd: mailwatcher.GetAllMailboxes,
			Params: map[string]interface{}{
				"emails": json.Marshal(),
			},
		}, nil
	}

	return nil, nil
}

func map2Mailbox(mp *map[string]interface{}) (*mailwatcher.Mailbox, error) {
	if mp == nil {
		return nil, errors.New("nil pointer to map")
	}

	errTemplate := "%s not in msg params"
	em, ok := (*mp)["email"].(string)
	if !ok {
		return nil, fmt.Errorf(errTemplate, "email")
	}
	pw, ok := (*mp)["password"].(string)
	if !ok {
		return nil, fmt.Errorf(errTemplate, "password")
	}
	srv, ok := (*mp)["server"].(string)
	if !ok {
		return nil, fmt.Errorf(errTemplate, "server")
	}
	port, ok := (*mp)["port"].(int32)
	if !ok {
		return nil, fmt.Errorf(errTemplate, "port")
	}
	useSSL, ok := (*mp)["useSSL"].(bool)
	if !ok {
		return nil, fmt.Errorf(errTemplate, "port")
	}
	mb := mailwatcher.Mailbox{
		Email:    em,
		Password: pw,
		Server:   srv,
		Port:     port,
		UseSSL:   useSSL,
	}
	return &mb, nil
}

func mailbox2map(mb *mailwatcher.Mailbox) map[string]interface{} {
	return map[string]interface{}{
		"email":  mb.Email,
		"server": mb.Server,
		"port":   mb.Port,
		"useSSL": mb.UseSSL,
	}
}
