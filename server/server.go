package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Server struct {
	listener         net.Listener
	quit             chan struct{}
	exited           chan struct{}
	db               memoryDB
	connections      map[int]net.Conn
	connCloseTimeout time.Duration
}

func NewServer() *Server {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("failed to create listener", err.Error())
	}
	srv := &Server{
		listener:         l,
		quit:             make(chan struct{}),
		exited:           make(chan struct{}),
		db:               newDB(),
		connections:      map[int]net.Conn{},
		connCloseTimeout: 10 * time.Second,
	}
	go srv.serve()
	return srv
}

func (srv *Server) serve() {
	var id int
	fmt.Println("listening for clients")
	for {
		select {
		case <-srv.quit:
			fmt.Println("shutting down the database server")
			err := srv.listener.Close()
			if err != nil {
				fmt.Println("could not close listener", err.Error())
			}
			if len(srv.connections) > 0 {
				srv.warnConnections(srv.connCloseTimeout)
				<-time.After(srv.connCloseTimeout)
				srv.closeConnections()
			}
			close(srv.exited)
			return
		default:
			tcpListener := srv.listener.(*net.TCPListener)
			err := tcpListener.SetDeadline(time.Now().Add(2 * time.Second))
			if err != nil {
				fmt.Println("failed to set listener deadline", err.Error())
			}

			conn, err := tcpListener.Accept()
			if oppErr, ok := err.(*net.OpError); ok && oppErr.Timeout() {
				continue
			}
			if err != nil {
				fmt.Println("failed to accept connection", err.Error())
			}

			write(conn, "Welcome to MemoryDB server")
			srv.connections[id] = conn
			go func(connID int) {
				fmt.Println("client with id", connID, "joined")
				srv.handleConn(conn)
				delete(srv.connections, connID)
				fmt.Println("client with id", connID, "left")
			}(id)
			id++
		}
	}
}

func write(conn net.Conn, s string) {
	_, err := fmt.Fprintf(conn, "%s\n-> ", s)
	if err != nil {
		log.Fatal(err)
	}
}

func (srv *Server) handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		l := strings.ToLower(strings.TrimSpace(scanner.Text()))
		values := strings.Split(l, " ")

		switch {
		case len(values) == 3 && values[0] == "set":
			srv.db.set(values[1], values[2])
			write(conn, "OK")
		case len(values) == 2 && values[0] == "get":
			k := values[1]
			val, found := srv.db.get(k)
			if !found {
				write(conn, fmt.Sprintf("key %s not found", k))
			} else {
				write(conn, val)
			}
		case len(values) == 2 && values[0] == "delete":
			srv.db.delete(values[1])
			write(conn, "OK")

		case len(values) == 1 && values[0] == "exit":
			if err := conn.Close(); err != nil {
				fmt.Println("could not close connection", err.Error())
			}
		default:
			write(conn, fmt.Sprintf("UNKNOWN: %s", l))
		}
	}
}

func (srv *Server) warnConnections(timeout time.Duration) {
	for _, conn := range srv.connections {
		write(conn, fmt.Sprintf("host wants to shut down the server in: %s", timeout.String()))
	}
}

func (srv *Server) closeConnections() {
	fmt.Println("closing all connections")
	for id, conn := range srv.connections {
		err := conn.Close()
		if err != nil {
			fmt.Println("could not close connection with id:", id)
		}
	}
}

func (srv *Server) Stop() {
	fmt.Println("stopping the database server")
	close(srv.quit)
	<-srv.exited
	fmt.Println("saving in-memory db records")
	srv.db.save()
	fmt.Println("database server successfully stopped")
}
