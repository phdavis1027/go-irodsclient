package testcases

import (
	"testing"

	"github.com/phdavis1027/go-irodsclient/irods/common"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("test ErrorString", testErrorString)
}

func testErrorString(t *testing.T) {
	errcode := common.REMOTE_SERVER_AUTHENTICATION_FAILURE

	// test - value
	errstr := common.GetIRODSErrorString(errcode)
	assert.Contains(t, errstr, "REMOTE_SERVER_AUTHENTICATION_FAILURE")

	// test + value
	errstr = common.GetIRODSErrorString(common.ErrorCode(-1 * int(errcode)))
	assert.Contains(t, errstr, "REMOTE_SERVER_AUTHENTICATION_FAILURE")

	// test sub value
	errcode = common.ErrorCode(int(common.REMOTE_SERVER_AUTHENTICATION_FAILURE) - int(common.EIO))
	assert.Equal(t, int(errcode), -910005)

	mainErrcode, subErrcode := common.SplitIRODSErrorCode(errcode)
	assert.Equal(t, common.REMOTE_SERVER_AUTHENTICATION_FAILURE, mainErrcode)
	assert.Equal(t, -1*common.EIO, subErrcode)

	errstr = common.GetIRODSErrorString(errcode)
	assert.Contains(t, errstr, "REMOTE_SERVER_AUTHENTICATION_FAILURE")
	assert.Contains(t, errstr, "I/O error")

}
