package filesender

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Relay interface {
	Start() error
	Stop()
}

func NewRelay(address string) (Relay, error) {
	return &relay{address: address, senders: make(map[string]chan handler)}, nil
}

type relay struct {
	address string
	sMutex  sync.Mutex
	senders map[string]chan handler // map secretKey receiverChan
}

func (r relay) registerSender(secretKey string, receiverChan chan handler) {
	r.sMutex.Lock()
	r.senders[secretKey] = receiverChan
	r.sMutex.Unlock()
}

func (r relay) takeSender(secretKey string) (chan handler, bool) {
	r.sMutex.Lock()
	defer r.sMutex.Unlock()

	h, ok := r.senders[secretKey]
	if !ok {
		return nil, false
	}
	delete(r.senders, secretKey)
	return h, ok
}

func (r relay) Start() error {
	s, err := net.Listen("tcp", r.address)
	if err != nil {
		return err
	}
	defer s.Close()

	for {
		connection, err := s.Accept()
		if err != nil {
			log.Println("Error: ", err)
		}
		go r.handleConnection(connection)
	}
}

func (r relay) Stop() {
	// todo
}

func (r relay) handleConnection(conn net.Conn) {
	var err error
	h := newHandler(conn)
	defer func() {
		if err != nil {
			h.Close()
		}
	}()

	cmd, registrationData, err := h.receiveRegisterCmd()
	if err != nil {
		return
	}

	switch cmd {
	case MSGRegisterSender:
		err = r.handleSender(h)
	case MSGRegisterReceiver:
		err = r.handleReceiver(registrationData, h)
	default:
		err = errors.New("wrong client type")
	}

	if err != nil {
		log.Println(err)
	}
	return
}

func (r relay) handleSender(h handler) error {
	secretKey := uuid.New().String()

	receiverChan := make(chan handler)
	r.registerSender(secretKey, receiverChan)
	err := h.sendRegisterSenderReply(secretKey)
	if err != nil {
		return err
	}

	timer := time.NewTimer(time.Hour * 12) // todo use conn DeadLine and context instead
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			r.takeSender(secretKey)
			return errors.New("timeout")
		case rh := <-receiverChan:
			proxying(h, rh)
			// one of the client disconnected, closing
			h.Close()
			rh.Close()
			// todo
			//case <- done:
			//	return nil
		}
	}
}

func (r relay) handleReceiver(secretKey string, h handler) error {
	receiverChan, ok := r.takeSender(secretKey)
	if !ok {
		h.Close()
		return errors.New("no sender found")
	}
	err := h.sendRegisterReceiverReply()
	if err != nil {
		return err
	}

	receiverChan <- h
	return nil
}

// todo shutdown
func proxying(senderConn, receiverConn io.ReadWriter) {

	receiverChan := chanFromConn(receiverConn)
	senderChan := chanFromConn(senderConn)

loop:
	for {
		select {
		case b1 := <-receiverChan:
			if b1 == nil {
				break loop
			} else {
				senderConn.Write(b1)
			}
		case b2 := <-senderChan:
			if b2 == nil {
				break loop
			} else {
				receiverConn.Write(b2)
			}
		}
	}
}

// chanFromConn creates a channel from a Conn object, and sends everything it
//  Read()s from the socket to the channel.
func chanFromConn(conn io.Reader) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				// Copy the buffer so it doesn't get changed while read by the recipient.
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				c <- nil
				break
			}
		}
	}()

	return c
}
