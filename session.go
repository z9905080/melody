package melody

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// Session wrapper around websocket connections.
type Session struct {
	Request  *http.Request
	Keys     map[string]interface{}
	keymutex *sync.RWMutex
	conn     net.Conn
	output   chan *envelope
	melody   *Melody
	open     bool
	rwmutex  *sync.RWMutex
	subChan  chan *envelope
}

// AddSub 訂閱某個,多個topic
func (s *Session) AddSub(topicNames ...string) {
	if s.subChan != nil {
		s.melody.pubsub.AddSub(s.subChan, topicNames...)
	} else {
		var str = ""
		for _, topicName := range topicNames {
			str += topicName + ","
		}
		str = str[0 : len(str)-1]
		s.melody.errorHandler(s, errors.New("error of add current channel,"+str))
	}
}

func (s *Session) writeMessage(message *envelope) {
	if s.closed() {
		s.melody.errorHandler(s, errors.New("tried to write to closed a session"))
		return
	}

	select {
	case s.output <- message:
	default:
		s.melody.errorHandler(s, errors.New("session message buffer is full"))
	}
}

func (s *Session) writeRaw(message *envelope) error {
	if s.closed() {
		return errors.New("tried to write to a closed session")
	}

	s.conn.SetWriteDeadline(time.Now().Add(s.melody.Config.WriteWait))
	err := wsutil.WriteServerMessage(s.conn, message.opCode, message.msg)
	//s.conn.WriteMessage(message.t, message.msg)

	if err != nil {
		return err
	}

	return nil
}

func (s *Session) closed() bool {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()

	return !s.open
}

func (s *Session) close() {
	if !s.closed() {
		s.rwmutex.Lock()
		s.open = false
		s.conn.Close()
		close(s.output)
		s.melody.pubsub.Unsub(s.subChan)
		s.rwmutex.Unlock()
	}
}

func (s *Session) ping() {
	s.writeRaw(&envelope{opCode: ws.OpPing, msg: []byte{}})
}

func (s *Session) writePump() {
	ticker := time.NewTicker(s.melody.Config.PingPeriod)
	defer ticker.Stop()

loop:
	for {
		select {
		case msg, ok := <-s.subChan:
			if !ok {
				break loop
			}
			err := s.writeRaw(msg)
			if err != nil {
				s.melody.errorHandler(s, err)
				break loop
			}

			if msg.opCode == ws.OpClose {
				break loop
			}

			if msg.opCode == ws.OpText {
				s.melody.messageSentHandler(s, msg.msg)
			}

			if msg.opCode == ws.OpText {
				s.melody.messageSentHandlerBinary(s, msg.msg)
			}

		case msg, ok := <-s.output:
			if !ok {
				break loop
			}

			err := s.writeRaw(msg)

			if err != nil {
				s.melody.errorHandler(s, err)
				break loop
			}

			if msg.opCode == ws.OpClose {
				break loop
			}

			if msg.opCode == ws.OpText {
				s.melody.messageSentHandler(s, msg.msg)
			}

			if msg.opCode == ws.OpText {
				s.melody.messageSentHandlerBinary(s, msg.msg)
			}
		case <-ticker.C:
			//wsutil.WriteServerMessage(s.conn, ws.OpPing, []byte(""))
			s.ping()
		}
	}
}

func (s *Session) readPump() {
	//s.conn.SetReadLimit(s.melody.Config.MaxMessageSize)
	s.conn.SetDeadline(time.Time{})
	// s.conn.SetPongHandler(func(string) error {
	// 	s.conn.SetReadDeadline(time.Now().Add(s.melody.Config.PongWait))
	// 	s.melody.pongHandler(s)
	// 	return nil
	// })

	// if s.melody.closeHandler != nil {
	// 	s.conn.SetCloseHandler(func(code int, text string) error {
	// 		return s.melody.closeHandler(s, code, text)
	// 	})
	// }

	for {
		if msg, opCode, err := wsutil.ReadClientData(s.conn); err != nil {
			break
			s.conn.Close()
		} else {
			//t, message, err := s.conn.ReadMessage()
			if err != nil {
				s.melody.errorHandler(s, err)
				break
			}

			if opCode == ws.OpText {
				s.melody.messageHandler(s, msg)
			}

			if opCode == ws.OpBinary {
				s.melody.messageHandlerBinary(s, msg)
			}
		}
	}
}

// Write writes message to session.
func (s *Session) Write(msg []byte) error {
	if s.closed() {
		return errors.New("session is closed")
	}

	s.writeMessage(&envelope{opCode: ws.OpText, msg: msg})

	return nil
}

// WriteBinary writes a binary message to session.
func (s *Session) WriteBinary(msg []byte) error {
	if s.closed() {
		return errors.New("session is closed")
	}

	s.writeMessage(&envelope{opCode: ws.OpBinary, msg: msg})

	return nil
}

// Close closes session.
func (s *Session) Close() error {
	if s.closed() {
		return errors.New("session is already closed")
	}

	s.writeMessage(&envelope{opCode: ws.OpClose, msg: []byte{}})

	return nil
}

// CloseWithMsg closes the session with the provided payload.
// Use the FormatCloseMessage function to format a proper close message payload.
func (s *Session) CloseWithMsg(msg []byte) error {
	if s.closed() {
		return errors.New("session is already closed")
	}

	s.writeMessage(&envelope{opCode: ws.OpClose, msg: msg})

	return nil
}

// Set is used to store a new key/value pair exclusivelly for this session.
// It also lazy initializes s.Keys if it was not used previously.
func (s *Session) Set(key string, value interface{}) {
	s.keymutex.Lock()
	defer s.keymutex.Unlock()
	if s.Keys == nil {
		s.Keys = make(map[string]interface{})
	}

	s.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (s *Session) Get(key string) (value interface{}, exists bool) {
	s.keymutex.RLock()
	defer s.keymutex.RUnlock()
	if s.Keys != nil {
		value, exists = s.Keys[key]
	}

	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (s *Session) MustGet(key string) interface{} {
	if value, exists := s.Get(key); exists {
		return value
	}

	panic("Key \"" + key + "\" does not exist")
}

// IsClosed returns the status of the connection.
func (s *Session) IsClosed() bool {
	return s.closed()
}
