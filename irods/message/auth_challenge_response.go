package message

import (
	"encoding/base64"
	"encoding/xml"

	"github.com/phdavis1027/go-irodsclient/irods/common"
	"golang.org/x/xerrors"
)

// IRODSMessageAuthChallengeResponse stores auth challenge
type IRODSMessageAuthChallengeResponse struct {
	XMLName   xml.Name `xml:"authRequestOut_PI"`
	Challenge string   `xml:"challenge"`
}

// GetBytes returns byte array
func (msg *IRODSMessageAuthChallengeResponse) GetBytes() ([]byte, error) {
	xmlBytes, err := xml.Marshal(msg)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal irods message to xml: %w", err)
	}
	return xmlBytes, nil
}

// FromBytes returns struct from bytes
func (msg *IRODSMessageAuthChallengeResponse) FromBytes(bytes []byte) error {
	err := xml.Unmarshal(bytes, msg)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal xml to irods message: %w", err)
	}
	return nil
}

// GetMessage builds a message
func (msg *IRODSMessageAuthChallengeResponse) GetMessage() (*IRODSMessage, error) {
	bytes, err := msg.GetBytes()
	if err != nil {
		return nil, xerrors.Errorf("failed to get bytes from irods message: %w", err)
	}

	msgBody := IRODSMessageBody{
		Type:    RODS_MESSAGE_API_REPLY_TYPE,
		Message: bytes,
		Error:   nil,
		Bs:      nil,
		IntInfo: int32(common.AUTH_REQUEST_AN),
	}

	msgHeader, err := msgBody.BuildHeader()
	if err != nil {
		return nil, xerrors.Errorf("failed to build header from irods message: %w", err)
	}

	return &IRODSMessage{
		Header: msgHeader,
		Body:   &msgBody,
	}, nil
}

// FromMessage returns struct from IRODSMessage
func (msg *IRODSMessageAuthChallengeResponse) FromMessage(msgIn *IRODSMessage) error {
	if msgIn.Body == nil {
		return xerrors.Errorf("empty message body")
	}

	err := msg.FromBytes(msgIn.Body.Message)
	if err != nil {
		return xerrors.Errorf("failed to get irods message from message body")
	}
	return nil
}

// GetChallenge returns challenge bytes
func (msg *IRODSMessageAuthChallengeResponse) GetChallenge() ([]byte, error) {
	challengeBytes, err := base64.StdEncoding.DecodeString(msg.Challenge)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode an authentication challenge: %w", err)
	}

	return challengeBytes, nil
}
