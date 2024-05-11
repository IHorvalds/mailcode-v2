package mailwatcher

import (
	"container/list"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type EmailCode struct {
	Sender string
	Code   string
}

type MailboxContext struct {
	mailbox     *Mailbox
	doneChannel chan struct{}
	isRunning   bool
	rwMtx       sync.RWMutex
}

func WatchMailboxes(mailboxes *list.List, config *Configuration, codeChannel chan EmailCode) (map[string]*MailboxContext, error) {
	contexts := map[string]*MailboxContext{}

	for e := mailboxes.Front(); e != nil; e = e.Next() {
		mb := e.Value.(*Mailbox)
		contexts[mb.Email] = new(MailboxContext)
		contexts[mb.Email].mailbox = mb
		contexts[mb.Email].doneChannel = make(chan struct{})
		contexts[mb.Email].isRunning = false
		contexts[mb.Email].rwMtx = sync.RWMutex{}

		go watchMailbox(contexts[mb.Email], config, codeChannel)
	}

	return contexts, nil
}

func WatchMailbox(mb *Mailbox, config *Configuration, codeChannel chan EmailCode) *MailboxContext {
	ctx := new(MailboxContext)
	ctx.mailbox = mb
	ctx.doneChannel = make(chan struct{})
	ctx.isRunning = false

	go watchMailbox(ctx, config, codeChannel)
	return ctx
}

func StopWatchingMailboxes(ctxs *map[string]*MailboxContext) {
	for _, ctx := range *ctxs {
		if IsRunning(ctx) {
			StopWatchingMailbox(ctx)
		}
	}
}

func StopWatchingMailbox(ctx *MailboxContext) {
	log.Printf("Stopping %s", ctx.mailbox.Email)
	ctx.doneChannel <- struct{}{}
}

func IsRunning(ctx *MailboxContext) bool {
	ctx.rwMtx.RLock()
	running := ctx.isRunning
	ctx.rwMtx.RUnlock()
	return running
}

func watchMailbox(ctx *MailboxContext, config *Configuration, codeChannel chan EmailCode) {
	var c *client.Client = nil
	var err error = nil
	if ctx.mailbox.UseSSL {
		c, err = client.DialTLS(fmt.Sprintf("%s:%d", ctx.mailbox.Server, ctx.mailbox.Port), nil)
	} else {
		c, err = client.Dial(fmt.Sprintf("%s:%d", ctx.mailbox.Server, ctx.mailbox.Port))
	}

	if err != nil {
		log.Println(err)
		return
	}

	err = c.Login(ctx.mailbox.Email, ctx.mailbox.Password)
	defer c.Logout()
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Starting to watch %s...\n", ctx.mailbox.Email)

	ctx.rwMtx.Lock()
	ctx.isRunning = true
	ctx.rwMtx.Unlock()

	defer func() {
		ctx.rwMtx.Lock()
		ctx.isRunning = false
		log.Printf("Stopped watching %s.\n", ctx.mailbox.Email)
		ctx.rwMtx.Unlock()
	}()

	if _, err = c.Select("INBOX", false); err != nil {
		log.Println(err)
		return
	}

	updates := make(chan client.Update)
	paused := make(chan error, 1)

	c.Updates = updates

	for {
		stop := make(chan struct{})
		go func() {
			paused <- c.Idle(stop, nil)
		}()

		stopped := false
		for {
			finishedIdling := false
			select {
			case <-ctx.doneChannel:
				close(stop)
				return
			case err = <-paused:
				if err != nil {
					log.Println(err)
					return
				}
				finishedIdling = true
			case <-updates:
				if !stopped {
					stop <- struct{}{}
					close(stop)
					stopped = true
				}
			}

			if finishedIdling {
				break
			}
		}

		messages := make(chan *imap.Message)
		done := make(chan error, 1)
		go func() {
			done <- fetchEmails(c, &config.Subjects, messages)
		}()

		seenUids := new(imap.SeqSet)
		for msg := range messages {
			code, extractErr := extractCode(msg, &config.Extractors)
			if extractErr != nil {
				log.Println(extractErr)
			} else {
				codeChannel <- code
				seenUids.AddNum(msg.SeqNum)
			}
		}

		if err := <-done; err != nil {
			log.Println(err)
			return
		}

		if !seenUids.Empty() {
			item := imap.FormatFlagsOp(imap.AddFlags, true)
			flags := []interface{}{imap.SeenFlag}
			errChan := make(chan error, 1)
			go func() {
				errChan <- c.Store(seenUids, item, flags, nil)
			}()
			if err := <-errChan; err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func fetchEmails(c *client.Client, subjects *[]string, messages chan *imap.Message) error {
	now := time.Now()
	criteria := imap.NewSearchCriteria()
	criteria.Since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Add(time.Duration(-24) * time.Hour)
	criteria.WithoutFlags = []string{imap.SeenFlag}

	if len(*subjects) == 1 {
		criteria.Header.Add("SUBJECT", (*subjects)[0])
	} else if len(*subjects) > 1 {
		subjectSearchCrit := new([2]*imap.SearchCriteria)
		criteria.Or = append(criteria.Or, *subjectSearchCrit)
		subjectSearchCrit = &(criteria.Or[0])

		for idx, sub := range *subjects {
			crit := imap.NewSearchCriteria()
			crit.Header.Add("SUBJECT", sub)
			if idx == len(*subjects)-1 {
				subjectSearchCrit[1] = crit
			} else {
				subjectSearchCrit[0] = crit
				if idx < len(*subjects)-2 {
					subjectSearchCrit[1] = imap.NewSearchCriteria()
					subjectSearchCrit[1].Or = [][2]*imap.SearchCriteria{}
					subjectSearchCrit[1].Or = append(subjectSearchCrit[1].Or, [2]*imap.SearchCriteria{})
					subjectSearchCrit = &subjectSearchCrit[1].Or[0]
				}
			}

			idx++
		}
	}

	uids, err := c.UidSearch(criteria)
	if err != nil {
		return err
	}

	if len(uids) > 0 {
		log.Println("Found potential verification code email")
		seqset := new(imap.SeqSet)
		seqset.AddNum(uids...)
		return c.UidFetch(seqset, []imap.FetchItem{imap.FetchItem("BODY.PEEK[TEXT] INTERNALDATE"), imap.FetchEnvelope}, messages)
	}

	close(messages)
	return nil
}

func extractCode(msg *imap.Message, regs *[]Extractor) (EmailCode, error) {
	c := EmailCode{}

	for _, literal := range msg.Body {
		body, err := io.ReadAll(literal)
		if err != nil {
			return c, err
		}

		sender := ""
		if len(msg.Envelope.From) > 0 {
			sender = msg.Envelope.From[0].Address()
		}

		strBody := string(body)
		for _, re := range *regs {
			parts := (&re.Reg).FindStringSubmatch(strBody)
			if capName, ok := re.Capture.(string); ok {
				capIdx := (&re.Reg).SubexpIndex(capName)
				if capIdx > -1 && len(parts) > capIdx {
					return EmailCode{
						Sender: sender,
						Code:   parts[capIdx],
					}, nil
				}
			} else if v, ok := re.Capture.(int); ok {
				if len(parts) > v {
					return EmailCode{
						Sender: sender,
						Code:   parts[v],
					}, nil
				}
			} else {
				log.Panicf("Unexpected type of capture: %s", reflect.TypeOf(re.Capture).Name())
			}
		}

		return c, errors.New("no codes found")
	}
	return c, errors.New("message had no body")
}
