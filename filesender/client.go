package filesender

import (
	"io"
)

type client struct {
	h         handler
	secretKey string
}

func (c *client) AskForFile() error {
	return c.h.sendStartFileTransfer()
}

func (c *client) SecretKey() string {
	return c.secretKey
}

func (c *client) WaitForProxy() error {
	return c.h.receiveStartFileTransfer()
}

func (c *client) SendFile(filename string, fileSize int64, file io.Reader) (int64, error) {

	err := c.h.sendFileHeader(filename)
	if err != nil {
		return 0, err
	}

	return c.h.sendFileData(fileSize, file)
}

func (c *client) ReceiveFile() (fileName string, fileSize int64, rc io.Reader, err error) {

	fileName, err = c.h.receiveFileHeader()
	if err != nil {
		return
	}

	fileSize, rc, err = c.h.receiveFileData()
	if err != nil {
		return
	}

	return fileName, fileSize, io.LimitReader(rc, fileSize), nil
}

func (c *client) Quit() error {
	return c.h.Close()
}
