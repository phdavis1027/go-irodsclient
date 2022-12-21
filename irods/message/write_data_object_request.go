package message

import (
	"encoding/xml"

	"github.com/cyverse/go-irodsclient/irods/common"
)

// IRODSMessageWriteDataObjectRequest stores data object read request
// type IRODSMessageWriteDataObjectRequest IRODSMessageOpenedDataObjectRequest
type IRODSMessageWriteDataObjectRequest struct {
	IRODSMessageOpenedDataObjectRequest
	Data []byte `xml:"-"`
}

// NewIRODSMessageWriteDataObjectRequest creates a IRODSMessageWriteDataObjectRequest message
func NewIRODSMessageWriteDataObjectRequest(desc int, data []byte) *IRODSMessageWriteDataObjectRequest {
	request := &IRODSMessageWriteDataObjectRequest{
		IRODSMessageOpenedDataObjectRequest: IRODSMessageOpenedDataObjectRequest{
			FileDescriptor: desc,
			Size:           int64(len(data)),
			Whence:         0,
			OperationType:  0,
			Offset:         0,
			BytesWritten:   0,
			KeyVals: IRODSMessageSSKeyVal{
				Length: 0,
			},
		},
		Data: data,
	}

	return request
}

// AddKeyVal adds a key-value pair
func (msg *IRODSMessageWriteDataObjectRequest) AddKeyVal(key common.KeyWord, val string) {
	msg.KeyVals.Add(string(key), val)
}

// GetBytes returns byte array
func (msg *IRODSMessageWriteDataObjectRequest) GetBytes() ([]byte, error) {
	xmlBytes, err := xml.Marshal(msg)
	return xmlBytes, err
}

// FromBytes returns struct from bytes
func (msg *IRODSMessageWriteDataObjectRequest) FromBytes(bytes []byte) error {
	err := xml.Unmarshal(bytes, msg)
	return err
}

// GetMessage builds a message
func (msg *IRODSMessageWriteDataObjectRequest) GetMessage() (*IRODSMessage, error) {
	bytes, err := msg.GetBytes()
	if err != nil {
		return nil, err
	}

	msgBody := IRODSMessageBody{
		Type:    RODS_MESSAGE_API_REQ_TYPE,
		Message: bytes,
		Error:   nil,
		Bs:      msg.Data,
		IntInfo: int32(common.DATA_OBJ_WRITE_AN),
	}

	msgHeader, err := msgBody.BuildHeader()
	if err != nil {
		return nil, err
	}

	return &IRODSMessage{
		Header: msgHeader,
		Body:   &msgBody,
	}, nil
}