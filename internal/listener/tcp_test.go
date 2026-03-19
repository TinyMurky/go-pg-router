package listener_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"

	tcplistener "github.com/TinyMurky/go-pg-router/internal/listener"
)

type EchoHandler struct {
	buf     bytes.Buffer
	writeCh chan<- string
}

var _ tcplistener.Handler = (*EchoHandler)(nil)

func NewEchoHandler(writeCh chan<- string) *EchoHandler {
	return &EchoHandler{
		writeCh: writeCh,
	}
}

func (eh *EchoHandler) Handle(conn net.Conn) {
	// ReadFrom will read until EOF
	// conn should be closed so that it send EOF
	eh.buf.ReadFrom(conn)
	eh.writeCh <- eh.buf.String()
}

func (eh *EchoHandler) String() string {
	return eh.buf.String()
}

type ListenerTestSuite struct {
	suite.Suite
}

func (suite *ListenerTestSuite) SetupTest() {}

func (suite *ListenerTestSuite) TestListenerHandleConn() {
	assert := suite.Assert()

	ch := make(chan string, 1)
	h := NewEchoHandler(ch)
	tcpListener := tcplistener.New(h)

	want := "echo"
	wantByte := []byte(want)

	l, err := net.Listen("tcp", "localhost:0")
	assert.NoError(err, "Expect no error when starting a tcp listen to localhost:0")
	defer l.Close()

	go func() {
		_ = tcpListener.Start(l)
	}()

	conn, err := net.Dial("tcp", l.Addr().String())

	assert.NoError(err, "Expect no error when dial to tcp listen")

	_, err = conn.Write(wantByte)
	assert.NoError(err, "Expect no error when writing to conn")
	conn.Close() // send EOF, so that ReadFrom can return

	got := <-ch
	assert.Equal(want, got)
}

func TestListenerTestSuite(t *testing.T) {
	suite.Run(t, new(ListenerTestSuite))
}
