package filesender

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

// probably unnecessary interface, could just use struct
type handler interface {
	sendRegisterSender() error
	sendRegisterReceiver(secretKey string) error
	receiveRegisterCmd() (cmd MessageType, msg string, err error)

	sendRegisterSenderReply(secretKey string) error
	receiveRegisterSenderReply() (secretKey string, err error)

	sendRegisterReceiverReply() error
	receiveRegisterReceiverReply() error

	sendStartFileTransfer() error
	receiveStartFileTransfer() error

	sendFileHeader(filename string) error
	receiveFileHeader() (filename string, err error)

	sendFileData(fileSize int64, r io.Reader) (n int64, err error)
	receiveFileData() (fileSize int64, r io.Reader, err error)

	io.Reader
	io.Writer
	io.Closer
}

type connHandler struct {
	conn net.Conn
}

func newHandler(conn net.Conn) handler {
	return &connHandler{conn: conn}
}

func (h *connHandler) Read(p []byte) (n int, err error) {
	return h.conn.Read(p)
}

func (h *connHandler) Write(p []byte) (n int, err error) {
	return h.conn.Write(p)
}

func (h *connHandler) Close() error {
	return h.conn.Close()
}

func (h *connHandler) sendRegisterReceiver(secretKey string) error {
	return h.sendMessage(MSGRegisterReceiver, secretKey)
}

func (h *connHandler) sendRegisterSender() error {
	return h.sendMessage(MSGRegisterSender, "")
}

func (h *connHandler) receiveRegisterCmd() (cmd MessageType, msg string, err error) {
	return h.receiveMessages(MSGRegisterReceiver, MSGRegisterSender)
}

func (h *connHandler) sendRegisterSenderReply(secretKey string) error {
	return h.sendMessage(MSGRegisterSenderReply, secretKey)
}

func (h *connHandler) receiveRegisterSenderReply() (secretKey string, err error) {
	return h.receiveMessage(MSGRegisterSenderReply)
}

func (h *connHandler) sendRegisterReceiverReply() error {
	return h.sendMessage(MSGRegisterReceiverReply, "")
}

func (h *connHandler) receiveRegisterReceiverReply() error {
	_, err := h.receiveMessage(MSGRegisterReceiverReply)
	return err
}

func (h *connHandler) sendStartFileTransfer() error {
	return h.sendMessage(MSGStartFileTransfer, "")
}

func (h *connHandler) receiveStartFileTransfer() error {
	_, err := h.receiveMessage(MSGStartFileTransfer)
	return err
}

func (h *connHandler) sendFileHeader(filename string) error {
	return h.sendMessage(MSGFileHeader, filename)
}

func (h *connHandler) receiveFileHeader() (filename string, err error) {
	return h.receiveMessage(MSGFileHeader)
}

func (h *connHandler) sendFileData(fileSize int64, r io.Reader) (n int64, err error) {
	return h.sendData(byte(MSGFileData), fileSize, r)
}

func (h *connHandler) receiveFileData() (fileSize int64, r io.Reader, err error) {
	cmd, datasize, r, err := h.receiveData()
	if err != nil {
		return 0, nil, err
	}

	if cmd != MSGFileData {
		return 0, nil, ErrWrongCmd
	}

	return datasize, io.LimitReader(r, datasize), nil
}

func (h *connHandler) sendMessage(cmd MessageType, msg string) error {
	buf := []byte(msg)
	dataSize := int64(len(buf))
	dataSizeLimit := msgLimits[cmd]

	if dataSizeLimit < dataSize {
		return ErrWrongDataSize
	}

	_, err := h.sendData(byte(cmd), dataSize, bytes.NewReader(buf))
	return err
}

func (h *connHandler) receiveMessage(cmd MessageType) (msg string, err error) {
	_, msg, err = h.receiveMessages(cmd)
	return msg, err
}

func (h *connHandler) receiveMessages(cmds ...MessageType) (cmd MessageType, msg string, err error) {
	cmd, datasize, r, err := h.receiveData()
	if err != nil {
		return 0, "", err
	}

	var found bool
	for _, c := range cmds {
		if c == cmd {
			found = true
			break
		}
	}
	if !found {
		return 0, "", ErrWrongCmd
	}

	if datasize == 0 {
		return cmd, "", nil
	}

	buf := make([]byte, datasize)
	_, err = r.Read(buf)
	if err != nil {
		return 0, "", err
	}

	return cmd, string(buf), nil
}

func (h *connHandler) sendData(cmd byte, dataSize int64, r io.Reader) (n int64, err error) {

	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(buf, dataSize)

	_, err = h.conn.Write(append([]byte{cmd}, buf...))
	if err != nil {
		return 0, err
	}

	if dataSize > 0 {
		n, err = io.CopyN(h.conn, r, dataSize)
		if err != nil {
			return 0, err
		}
	}
	return n, err
}

func (h *connHandler) receiveData() (cmd MessageType, dataSize int64, r io.Reader, err error) {
	buff := make([]byte, 11)
	_, err = h.conn.Read(buff)
	cmdByte := buff[0]
	dataSize, _ = binary.Varint(buff[1:])
	return MessageType(cmdByte), dataSize, h.conn, err
}
