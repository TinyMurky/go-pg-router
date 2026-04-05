package pgserver_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/TinyMurky/go-pg-router/internal/pgserver"
)

type StartupMessageTestSuite struct {
	suite.Suite
}

func (suite *StartupMessageTestSuite) SetupTest() {}

func (suite *StartupMessageTestSuite) TestReadStartupMessage_HappyPass() {
	t := suite.T()

	assert := suite.Assert()
	// golang use \000 as null in c
	wantKVs := map[string]string{
		"testA": "3",
		"testB": "ball",
	}

	testStartupMsg, wantProtocol := pgserver.BuildStarupMessage(t, wantKVs)

	startUpMessage := new(pgserver.StartupMessage)
	err := startUpMessage.ReadStartupMessage(bytes.NewReader(testStartupMsg))

	assert.NoError(err)

	assert.Equal(wantProtocol, startUpMessage.ProtocolVersion)
	assert.Equal(wantKVs, startUpMessage.Parameters)
}

func (suite *StartupMessageTestSuite) TestReadStartupMessage_InvalidMsgFormat() {
	assert := suite.Assert()

	testStartupMsg := []byte("invalid long long msg")

	startUpMessage := new(pgserver.StartupMessage)
	err := startUpMessage.ReadStartupMessage(bytes.NewReader(testStartupMsg))

	assert.ErrorIs(err, pgserver.ErrInvalidMsgFormat)
}

func (suite *StartupMessageTestSuite) TestReadStartupMessage_TotalLengthTooShort() {
	t := suite.T()

	assert := suite.Assert()

	kv := map[string]string{
		"testA": "3",
		"testB": "ball",
	}
	binaryTestKVs := pgserver.BuildBinaryKV(t, kv)

	var testProtocol uint32 = 0x00030000
	// total message lenght should be 4 bytes too
	var testTotalLength uint32 = 4 + 4 + 1

	msg := pgserver.BuildCustomStartupMsg(t, testTotalLength, testProtocol, binaryTestKVs)

	startUpMessage := new(pgserver.StartupMessage)
	err := startUpMessage.ReadStartupMessage(bytes.NewReader(msg))

	assert.ErrorIs(err, pgserver.ErrInvalidMsgFormat)
}

func (suite *StartupMessageTestSuite) TestReadStartupMessage_TotalLengthTooLong() {
	t := suite.T()

	assert := suite.Assert()

	kv := map[string]string{
		"testA": "3",
		"testB": "ball",
	}
	binaryTestKVs := pgserver.BuildBinaryKV(t, kv)

	var testProtocol uint32 = 0x00030000
	// total message lenght should be 4 bytes too
	var testTotalLength uint32 = 4 + 4 + 1000

	msg := pgserver.BuildCustomStartupMsg(t, testTotalLength, testProtocol, binaryTestKVs)

	startUpMessage := new(pgserver.StartupMessage)
	err := startUpMessage.ReadStartupMessage(bytes.NewReader(msg))

	assert.ErrorIs(err, pgserver.ErrConnectionClosed)
}

func (suite *StartupMessageTestSuite) TestReadStartupMessage_KVWrongFormat() {
	t := suite.T()

	assert := suite.Assert()

	binaryTestKVs := []byte("a\000\000") // 3 bytes

	var testProtocol uint32 = 0x00030000
	// total message lenght should be 4 bytes too
	var testTotalLength uint32 = 4 + 4 + 3

	msg := pgserver.BuildCustomStartupMsg(t, testTotalLength, testProtocol, binaryTestKVs)

	startUpMessage := new(pgserver.StartupMessage)
	err := startUpMessage.ReadStartupMessage(bytes.NewReader(msg))

	assert.ErrorIs(err, pgserver.ErrInvalidMsgFormat)
}

func (suite *StartupMessageTestSuite) TestWriteAuthOK() {

	assert := suite.Assert()
	want := pgserver.StartUPAuthenticationOk()

	var testBuf bytes.Buffer

	startUpMessage := new(pgserver.StartupMessage)

	err := startUpMessage.WriteAuthOK(&testBuf)
	assert.NoError(err)

	assert.Equal(want, testBuf.Bytes())
}

func (suite *StartupMessageTestSuite) TestWriteReadyForQuery() {

	assert := suite.Assert()
	want := pgserver.StartUPReadyForQuery()

	var testBuf bytes.Buffer

	startUpMessage := new(pgserver.StartupMessage)

	err := startUpMessage.WriteReadyForQuery(&testBuf)

	assert.NoError(err)

	assert.Equal(want, testBuf.Bytes())
}

func TestStartupMessageTestSuite(t *testing.T) {
	suite.Run(t, new(StartupMessageTestSuite))
}
