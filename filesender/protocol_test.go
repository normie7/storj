package filesender

import (
	"bytes"
	"net"
	"reflect"
	"testing"
	"time"
)

type tester struct {
	buffer   []byte
	deadline time.Time
}

func (c *tester) LocalAddr() net.Addr                { panic("implement me") }
func (c *tester) RemoteAddr() net.Addr               { panic("implement me") }
func (c *tester) SetDeadline(t time.Time) error      { panic("implement me") }
func (c *tester) SetReadDeadline(t time.Time) error  { panic("implement me") }
func (c *tester) SetWriteDeadline(t time.Time) error { panic("implement me") }
func (c *tester) Read(b []byte) (n int, err error) {
	c.buffer = c.buffer[copy(b, c.buffer):]
	return len(c.buffer), nil
}

func (c *tester) Write(b []byte) (n int, err error) {
	c.buffer = append(c.buffer, b...)
	return len(b), nil
}

func (c *tester) Close() error { return nil }

func TestSendData(t *testing.T) {
	testConn := &tester{}
	h := connHandler{conn: testConn}

	testBuffer := []byte("testing data sender")

	_, err := h.sendData(byte(MSGRegisterReceiver), int64(len(testBuffer)), bytes.NewReader(testBuffer))
	if err != nil {
		t.Fatal(err)
	}

	cmd, dataSize, r, err := h.receiveData()
	if err != nil {
		t.Fatal(err)
	}
	if cmd != MSGRegisterReceiver {
		t.Fatal("unexpected cmd")
	}
	if dataSize != int64(len(testBuffer)) {
		t.Fatal("unexpected dataSize")
	}

	receiverBuffer := make([]byte, dataSize)
	if _, err = r.Read(receiverBuffer); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(testBuffer, receiverBuffer) {
		t.Fatal("unexpected data")
	}

	if len(testConn.buffer) != 0 {
		t.Fatal("unread data left in the connection")
	}
}
