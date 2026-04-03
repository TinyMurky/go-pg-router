package pgproto_test

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/TinyMurky/go-pg-router/internal/pgproto"
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

	pgHandler := pgproto.NewPGHandler()

	client, server := net.Pipe()
	defer client.Close()

	go func() {
		defer server.Close()
		pgHandler.Handle(server)
	}()

	_, err := client.Write(testStartupMsg)
	assert.NoError(err)

	wantedAuthOK := make([]byte, len(pgproto.StartUPAuthenticationOk()))

	// use ReadFull so that it will return an error if Handler doesn't write enough bytes
	// (it will throw unexpected EOF)
	n, err := io.ReadFull(client, wantedAuthOK)
	assert.NoError(err)
	assert.Equal(len(pgproto.StartUPAuthenticationOk()), n)
	assert.Equal(pgproto.StartUPAuthenticationOk(), wantedAuthOK)

	wantedReadyForQuery := make([]byte, len(pgproto.StartUPReadyForQuery()))

	// use ReadFull so that it will return an error if Handler doesn't write enough bytes
	// (it will throw unexpected EOF)
	n, err = io.ReadFull(client, wantedReadyForQuery)
	assert.NoError(err)
	assert.Equal(len(pgproto.StartUPReadyForQuery()), n)
	assert.Equal(pgproto.StartUPReadyForQuery(), wantedReadyForQuery)
}

func (suite *PGHandlerTestSuite) TestHandle_Startup_ConnectCloseImediately() {
	// t := suite.T()
	assert := suite.Assert()

	pgHandler := pgproto.NewPGHandler()

	client, server := net.Pipe()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		pgHandler.Handle(server)

		// This make sure that server close after client.Close()
		<-ctx.Done()
		server.Close()
	}()

	// client close imediatly
	client.Close()

	buf := make([]byte, 1)
	_, err := server.Read(buf)

	// Makesure client close => read => server close
	cancel()

	// Server should be closed too
	assert.ErrorIs(err, io.EOF)

}

func TestPGHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PGHandlerTestSuite))
}
