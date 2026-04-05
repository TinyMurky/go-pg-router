package pgserver

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type SSLTestSuite struct {
	suite.Suite
}

func (suite *SSLTestSuite) SetupTest() {}

func (suite *SSLTestSuite) TestReedSSLPrStartup_StartWithSSLThenStartupMessage() {
	t := suite.T()
	assert := suite.Assert()

	client, server := net.Pipe()
	defer server.Close()

	testSSLRequest := BuildSSLRequestMessage(t)

	// golang use \000 as null in c
	testKVs := map[string]string{
		"testA": "3",
		"testB": "ball",
	}

	testStartupMsg, _ := BuildStarupMessage(t, testKVs)

	errCh := make(chan error, 3)
	go func() {

		// goroutine (client) ──writes SSL probe──► server (readSSLOrStartup reads it)
		// goroutine (client) ◄──writes 'N'──────── server (readSSLOrStartup writes it)

		defer client.Close()

		_, err := client.Write(testSSLRequest)
		errCh <- err

		nBuf := make([]byte, len(SSLRequestReject()))

		_, err = client.Read(nBuf)
		errCh <- err

		_, err = client.Write(testStartupMsg)
		errCh <- err

		close(errCh)
	}()

	got, err := readSSLOrStartup(server)
	assert.NoError(err)

	assert.Equal(testStartupMsg, got)

	// drain the error channel

	for err := range errCh {
		assert.NoError(err)
	}

}

func (suite *SSLTestSuite) TestReedSSLPrStartup_StartupMessageDirectly() {
	t := suite.T()
	assert := suite.Assert()

	client, server := net.Pipe()
	defer server.Close()

	// golang use \000 as null in c
	testKVs := map[string]string{
		"testA": "3",
		"testB": "ball",
	}

	testStartupMsg, _ := BuildStarupMessage(t, testKVs)

	errCh := make(chan error, 2)
	go func() {

		defer client.Close()

		// write StartupMessage without SSLRequest
		_, err := client.Write(testStartupMsg)
		errCh <- err

		close(errCh)
	}()

	got, err := readSSLOrStartup(server)
	assert.NoError(err)

	assert.Equal(testStartupMsg, got)

	// drain the error channel

	for err := range errCh {
		assert.NoError(err)
	}

}

func (suite *SSLTestSuite) TestReedSSLPrStartup_StartWithSSLThenCloseBeforeStartupMessage() {
	t := suite.T()
	assert := suite.Assert()

	client, server := net.Pipe()
	defer server.Close()

	testSSLRequest := BuildSSLRequestMessage(t)

	done := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {

		_, err := client.Write(testSSLRequest)
		errCh <- err

		// Closed After client send SSLRequest
		client.Close()

		close(errCh)
		done <- struct{}{}
	}()

	_, err := readSSLOrStartup(server)
	assert.ErrorIs(err, ErrConnectionClosed)
	select {
	case <-done:
		// pass test
	case <-time.After(time.Second):
		t.Fatal("goroutine did not return on time")
	}

	// drain the error channel
	for err := range errCh {
		assert.NoError(err)
	}

}

func (suite *SSLTestSuite) TestReedSSLPrStartup_StartWithInvalidSSLRequest() {
	t := suite.T()
	assert := suite.Assert()

	client, server := net.Pipe()

	testSSLRequest := []byte("Invalid SSL Request")
	done := make(chan struct{})
	go func() {
		client.Write(testSSLRequest)
		client.Close()
		close(done)
	}()

	_, err := readSSLOrStartup(server)
	// Close server right after error
	// so that the remain unread message of testSSLRequest
	// will not block
	server.Close()
	assert.ErrorIs(err, ErrInvalidMsgFormat)

	select {
	case <-done:
		// pass test
	case <-time.After(time.Second):
		t.Fatal("goroutine did not return on time")
	}
}

func (suite *SSLTestSuite) TestReedSSLPrStartup_StartWithSSLTwiceThenStartupMessage() {
	t := suite.T()
	assert := suite.Assert()

	client, server := net.Pipe()
	defer server.Close()

	testSSLRequest := BuildSSLRequestMessage(t)

	// golang use \000 as null in c
	testKVs := map[string]string{
		"testA": "3",
		"testB": "ball",
	}

	testStartupMsg, _ := BuildStarupMessage(t, testKVs)

	errCh := make(chan error, 5)
	go func() {

		// goroutine (client) ──writes SSL probe──► server (readSSLOrStartup reads it)
		// goroutine (client) ◄──writes 'N'──────── server (readSSLOrStartup writes it)

		defer client.Close()

		_, err := client.Write(testSSLRequest)
		errCh <- err

		nBuf1 := make([]byte, len(SSLRequestReject()))

		_, err = client.Read(nBuf1)
		errCh <- err

		_, err = client.Write(testSSLRequest)
		errCh <- err

		nBuf2 := make([]byte, len(SSLRequestReject()))

		_, err = client.Read(nBuf2)
		errCh <- err

		_, err = client.Write(testStartupMsg)
		errCh <- err

		close(errCh)
	}()

	got, err := readSSLOrStartup(server)
	assert.NoError(err)

	assert.Equal(testStartupMsg, got)

	// drain the error channel

	for err := range errCh {
		assert.NoError(err)
	}

}

func TestSSLMessageTestSuite(t *testing.T) {
	suite.Run(t, new(SSLTestSuite))
}
