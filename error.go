package melody

import "errors"

var (
	ErrWriteToCloseSessionForRecover = errors.New("tried to write to closed a session for recover")
	ErrWriteToCloseSession           = errors.New("tried to write to closed a session")
	ErrSessionMessageBufferIsFull    = errors.New("session message buffer is full")
)
