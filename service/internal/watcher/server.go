package watcher

import (
	"container/list"
	"errors"
	"io"
	"log"
	"mailcode/service/internal/mailwatcher"
	"net"
	"os"
	"sync"
)

type Server struct {
	watcher *Watcher

	connections *list.List
	mux         sync.Mutex

	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
}

func NewServer(path string) *Server {
	s := &Server{
		quit: make(chan interface{}),
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalln(err)
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		log.Fatalln(err)
	}
	s.listener = l
	s.mux = sync.Mutex{}
	s.connections = list.New().Init()

	s.wg.Add(1)
	return s
}

func (s *Server) Stop() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func (s *Server) Serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Println("accept error", err)
			}
		} else {
			s.wg.Add(1)

			s.mux.Lock()
			s.connections.PushBack(conn)
			s.mux.Unlock()

			go func() {
				s.HandleConection(conn)
				s.wg.Done()
			}()
		}
	}
}

func (s *Server) HandleConection(c net.Conn) {
	defer c.Close()
	buf := make([]byte, 2048)
	for {
		n, err := c.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("read error", err)
			return
		}
		if n == 0 {
			// Remove c from s.connections
			s.mux.Lock()
			for el := s.connections.Front(); el != nil; el = el.Next() {
				conn := el.Value.(net.Conn)
				if conn == c {
					s.connections.Remove(el)
				}
			}
			s.mux.Unlock()
			return
		}

		msg, err := mailwatcher.Parse(buf[:n])
		if err != nil {
			log.Println("Failed to parse received message")
		}

		s.watcher.handleMessage(&msg)
	}
}
