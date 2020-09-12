package filesender

import (
	"errors"
	"math"
)

const (
	MSGRegisterSender        MessageType = 10
	MSGRegisterReceiver      MessageType = 20
	MSGRegisterSenderReply   MessageType = 30
	MSGRegisterReceiverReply MessageType = 40
	MSGStartFileTransfer     MessageType = 50
	MSGFileHeader            MessageType = 60
	MSGFileData              MessageType = 255

	maxMSGSize = 1048576 // 1 MB
)

var (
	ErrCantDecode    = errors.New("can't decode")
	ErrWrongCmd      = errors.New("wrong command")
	ErrWrongDataSize = errors.New("wrong data size")

	// if MSG is not in this map it means that the limit is 0
	msgLimits = map[MessageType]int64{
		MSGRegisterReceiver:    maxMSGSize,
		MSGRegisterSenderReply: maxMSGSize,
		MSGFileHeader:          maxMSGSize,
		MSGFileData:            math.MaxInt64,
	}
)

type MessageType byte
