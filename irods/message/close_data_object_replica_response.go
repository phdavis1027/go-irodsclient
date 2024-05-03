package message

import (
	"github.com/phdavis1027/go-irodsclient/irods/common"
	"github.com/phdavis1027/go-irodsclient/irods/types"
	"golang.org/x/xerrors"
)

// IRODSMessageCloseDataObjectReplicaResponse stores data object replica close response
type IRODSMessageCloseDataObjectReplicaResponse struct {
	// empty structure
	Result int
}

// CheckError returns error if server returned an error
func (msg *IRODSMessageCloseDataObjectReplicaResponse) CheckError() error {
	if msg.Result < 0 {
		return types.NewIRODSError(common.ErrorCode(msg.Result))
	}
	return nil
}

// FromMessage returns struct from IRODSMessage
func (msg *IRODSMessageCloseDataObjectReplicaResponse) FromMessage(msgIn *IRODSMessage) error {
	if msgIn.Body == nil {
		return xerrors.Errorf("empty message body")
	}

	msg.Result = int(msgIn.Body.IntInfo)
	return nil
}
