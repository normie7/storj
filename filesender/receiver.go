package filesender

import (
	"io"
	"net"
)

type Receiver interface {
	AskForFile() error
	ReceiveFile() (fileName string, fileSize int64, rc io.Reader, err error)
	Quit() error
}

func RegisterReceiver(address, secretCode string) (Receiver, error) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	h := newHandler(conn)
	err = h.sendRegisterReceiver(secretCode)
	if err != nil {
		return nil, err
	}

	err = h.receiveRegisterReceiverReply()
	if err != nil {
		return nil, err
	}

	return &client{h: h}, nil
}
