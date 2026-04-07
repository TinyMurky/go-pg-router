package backend

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
}

func (suite *AuthTestSuite) SetupTest() {}

func (suite *AuthTestSuite) TestReadBackendMessage_ShouldParse() {
	t := suite.T()

	type testCase struct {
		name          string
		input         []byte
		expectMsgType byte
		expectPayload []byte
	}

	testCases := []testCase{
		{
			name:          "AuthenticationOk",
			input:         AuthenticationOk(),
			expectMsgType: byte('R'),
			expectPayload: []byte{
				// 0x00, 0x00, 0x00, 0x08,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name:          "AuthenticationCleartextPassword",
			input:         AuthenticationCleartextPassword(),
			expectMsgType: byte('R'),
			expectPayload: []byte{
				// 0x00, 0x00, 0x00, 0x08,
				0x00, 0x00, 0x00, 0x03,
			},
		},
		{
			name:          "AuthenticationMD5Password",
			input:         AuthenticationMD5Password(1),
			expectMsgType: byte('R'),
			expectPayload: []byte{
				// 0x00, 0x00, 0x00, 0x08,
				0x00, 0x00, 0x00, 0x05, // this is auth type
				0x00, 0x00, 0x00, 0x01, // salt is 1
			},
		},
		{
			name:          "ReadyForQuery",
			input:         ReadyForQuery(),
			expectMsgType: byte('Z'),
			expectPayload: []byte{
				// 0x00, 0x00, 0x00, 0x05, // this is length
				byte('I'),
			},
		},
		{
			name:          "ParameterStatus",
			input:         ParameterStatus("name", "value"),
			expectMsgType: byte('S'),
			expectPayload: []byte{
				0x6e, 0x61, 0x6d, 0x65, // name
				0x00,                         // \000
				0x76, 0x61, 0x6c, 0x75, 0x65, // value
				0x00, // \000

			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("should be able to read %s", tc.name), func(t *testing.T) {
			assert := require.New(t)
			gotMsgType, gotPayload, err := readBackendMessage(bytes.NewReader(tc.input))

			assert.NoError(err)
			assert.Equal(tc.expectMsgType, gotMsgType)
			assert.Equal(tc.expectPayload, gotPayload)
		})

	}
}

func (suite *AuthTestSuite) TestReadBackendMessage_InvalidMsgFormat() {
	t := suite.T()
	t.Run("total length is too long", func(t *testing.T) {
		assert := require.New(t)

		testMsg := buildCustomReadMessage(byte('R'), 1<<31, []byte("short message"))
		_, _, err := readBackendMessage(bytes.NewReader(testMsg))

		assert.ErrorIs(err, ErrInvalidMsgFormat)
	})
	t.Run("total length is too short", func(t *testing.T) {
		assert := require.New(t)

		testMsg := buildCustomReadMessage(byte('R'), 3, []byte{})
		_, _, err := readBackendMessage(bytes.NewReader(testMsg))

		assert.ErrorIs(err, ErrInvalidMsgFormat)
	})
	t.Run("payload was truncated(which means connect lost the following bytes)", func(t *testing.T) {
		assert := require.New(t)

		testMsg := buildCustomReadMessage(byte('R'), 999999, []byte("truncated message"))
		_, _, err := readBackendMessage(bytes.NewReader(testMsg))

		assert.ErrorIs(err, ErrConnectionClosed)
	})

	t.Run("input was truncated at mid-length (which means connect lost the following bytes)", func(t *testing.T) {
		assert := require.New(t)

		testMsg := buildCustomReadMessage(byte('R'), 5, []byte{0x01})
		_, _, err := readBackendMessage(bytes.NewReader(testMsg[:3]))

		assert.ErrorIs(err, ErrConnectionClosed)
	})

	t.Run("input truncated after msgType (no length bytes)", func(t *testing.T) {
		assert := require.New(t)

		testMsg := buildCustomReadMessage(byte('R'), 5, []byte{0x01})
		_, _, err := readBackendMessage(bytes.NewReader(testMsg[:1]))

		assert.ErrorIs(err, ErrConnectionClosed)
	})

}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
