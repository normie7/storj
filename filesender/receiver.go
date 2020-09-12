package filesender

import (
	"io"
	"net"
)

type Receiver struct {
	h *connHandler
}

func NewReceiver(address, secretCode string) (*Receiver, error) {

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

	return &Receiver{h: h}, nil
}

func (r *Receiver) AskForFile() error {
	return r.h.sendStartFileTransfer()
}

func (r *Receiver) ReceiveFile() (fileName string, fileSize int64, rc io.Reader, err error) {

	fileName, err = r.h.receiveFileHeader()
	if err != nil {
		return
	}

	fileSize, rc, err = r.h.receiveFileData()
	if err != nil {
		return
	}

	return fileName, fileSize, io.LimitReader(rc, fileSize), nil
}

func (r *Receiver) Quit() error {
	return r.h.Close()
}
