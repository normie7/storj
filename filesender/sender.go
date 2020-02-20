package filesender

import (
	"io"
	"net"
)

type Sender interface {
	SecretKey() string
	WaitForProxy() error
	SendFile(fileName string, fileSize int64, file io.Reader) (int64, error)
	Quit() error
}

func RegisterSender(address string) (Sender, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	h := newHandler(conn)
	err = h.sendRegisterSender()
	if err != nil {
		return nil, err
	}

	secretKey, err := h.receiveRegisterSenderReply()

	return &client{h: h, secretKey: secretKey}, err
}
