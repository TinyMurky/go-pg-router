package pgserver_test

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/TinyMurky/go-pg-router/internal/pgserver"
)

type PGHandlerTestSuite struct {
	suite.Suite
}

func (suite *PGHandlerTestSuite) SetupTest() {}

func (suite *PGHandlerTestSuite) TestHandle_Startup() {
	t := suite.T()
	assert := suite.Assert()
	// golang use \000 as null in c
	wantKVs := map[string]string{
		"testA": "3",
		"testB": "ball",
	}

	testStartupMsg, _ := BuildStarupMessage(t, wantKVs)

	pgHandler := pgserver.NewPGHandler()

	client, server := net.Pipe()
	defer client.Close()

	go func() {
		defer server.Close()
		pgHandler.Handle(server)
	}()

	_, err := client.Write(testStartupMsg)
	assert.NoError(err)

	wantedAuthOK := make([]byte, len(pgserver.StartUPAuthenticationOk()))

	// use ReadFull so that it will return an error if Handler doesn't write enough bytes
	// (it will throw unexpected EOF)
	n, err := io.ReadFull(client, wantedAuthOK)
	assert.NoError(err)
	assert.Equal(len(pgserver.StartUPAuthenticationOk()), n)
	assert.Equal(pgserver.StartUPAuthenticationOk(), wantedAuthOK)

	wantedReadyForQuery := make([]byte, len(pgserver.StartUPReadyForQuery()))

	// use ReadFull so that it will return an error if Handler doesn't write enough bytes
	// (it will throw unexpected EOF)
	n, err = io.ReadFull(client, wantedReadyForQuery)
	assert.NoError(err)
	assert.Equal(len(pgserver.StartUPReadyForQuery()), n)
	assert.Equal(pgserver.StartUPReadyForQuery(), wantedReadyForQuery)
}

func (suite *PGHandlerTestSuite) TestHandle_Startup_ConnectCloseImediately() {
	t := suite.T()

	pgHandler := pgserver.NewPGHandler()

	client, server := net.Pipe()

	done := make(chan struct{})

	go func() {
		defer server.Close()
		pgHandler.Handle(server)

		// signal that Handle returned, so the test can verify it didn't block
		done <- struct{}{}
	}()

	client.Close()

	select {
	case <-done:
		// pass test
	case <-time.After(time.Second):
		t.Fatal("goroutine did not return on time")
	}
}

func (suite *PGHandlerTestSuite) TestHandle_Startup_ConnectDropMidStartup() {
	t := suite.T()

	assert := suite.Assert()

	pgHandler := pgserver.NewPGHandler()

	client, server := net.Pipe()
	done := make(chan struct{})

	go func() {
		defer server.Close()
		pgHandler.Handle(server)

		// signal that Handle returned, so the test can verify it didn't block
		done <- struct{}{}
	}()

	// write 4 byte into server and close
	var testTotalLength uint32 = 10
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, testTotalLength)

	_, err := client.Write(buf)
	assert.NoError(err)

	client.Close()

	select {
	case <-done:
		// pass test
	case <-time.After(time.Second):
		t.Fatal("goroutine did not return on time")
	}
}

func TestPGHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PGHandlerTestSuite))
}
