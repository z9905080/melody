package melody

import "github.com/gobwas/ws"

type envelope struct {
	opCode ws.OpCode
	msg    []byte
	filter filterFunc
}
