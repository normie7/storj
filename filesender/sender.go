package filesender

import (
	"io"
	"net"
)

type Sender struct {
	h         *connHandler
	secretKey string
}

func NewSender(address string) (*Sender, error) {
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

	return &Sender{h: h, secretKey: secretKey}, err
}

func (s *Sender) SecretKey() string {
	return s.secretKey
}

func (s *Sender) WaitForProxy() error {
	return s.h.receiveStartFileTransfer()
}

func (s *Sender) SendFile(filename string, fileSize int64, file io.Reader) (int64, error) {

	err := s.h.sendFileHeader(filename)
	if err != nil {
		return 0, err
	}

	return s.h.sendFileData(fileSize, file)
}

func (s *Sender) Quit() error {
	return s.h.Close()
}
