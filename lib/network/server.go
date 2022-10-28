package network

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/kkevinchou/izzet/izzet/settings"
)

type Server struct {
	host           string
	port           string
	connectionType string

	nextID      int
	nextIDMutex sync.Mutex

	incomingConnections chan *Connection
}

func NewServer(host, port, connectionType string, idStart int) *Server {
	return &Server{
		host:           host,
		port:           port,
		connectionType: connectionType,

		nextID: idStart,

		incomingConnections: make(chan *Connection, incomingConnectionsBufferSize),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen(s.connectionType, s.host+":"+s.port)
	if err != nil {
		return err
	}
	fmt.Println("listening on " + s.host + ":" + s.port)

	if settings.LatencyInjection > 0 {
		listener = WrapListener(listener, settings.LatencyInjection)
	}

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("error accepting a connection on the listener:", err.Error())
				continue
			}

			id := s.generateNextID()

			message, err := s.createAcceptMessage(id)
			if err != nil {
				fmt.Println(err)
				continue
			}

			sendMessage(conn, message)
			if err != nil {
				fmt.Println("error sending accept message:", err.Error())
				continue
			}

			select {
			case s.incomingConnections <- &Connection{ID: id, Connection: conn}:
			default:
				panic("incomingConnections queue full")
			}
		}
	}()

	return nil
}

func (s *Server) PullIncomingConnections() []*Connection {
	connections := []*Connection{}

	for i := 0; i < incomingConnectionsBufferSize; i++ {
		select {
		case connection := <-s.incomingConnections:
			connections = append(connections, connection)
		default:
			return connections
		}
	}

	return connections
}

func sendMessage(connection net.Conn, message *Message) error {
	encoder := json.NewEncoder(connection)
	err := encoder.Encode(message)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) generateNextID() int {
	s.nextIDMutex.Lock()
	id := s.nextID
	s.nextID++
	s.nextIDMutex.Unlock()

	return id
}

func (s *Server) createAcceptMessage(id int) (*Message, error) {
	acceptMessage := AcceptMessage{
		ID: id,
	}
	bodyBytes, err := json.Marshal(acceptMessage)
	if err != nil {
		fmt.Println("error marshaling accept message:", err.Error())
		return nil, err
	}
	return &Message{
		SenderID:    0,
		MessageType: MessageTypeAcceptConnection,
		Body:        bodyBytes,
	}, nil

}
