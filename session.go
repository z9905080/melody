package melody

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Session wrapper around websocket connections.
type Session struct {
	Request         *http.Request
	Keys            map[string]interface{}
	keymutex        *sync.RWMutex
	conn            *websocket.Conn
	output          chan *envelope
	closeOutputChan chan struct{}
	melody          *Melody
	open            bool
	hashID          string
	rwmutex         *sync.RWMutex
	subChan         chan *envelope
}

// GetHashID 取得 HashID (Get Session HashID)
func (s *Session) GetHashID() string {
	return s.hashID
}

// AddSub 訂閱某個,多個topic (Session subscribe one or multi topics)
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

// UnSub (Session unsubscribe one or multi topics, if no topics ,will unsubscribe all topics)
func (s *Session) UnSub(topicNames ...string) {
	if s.subChan != nil {
		s.melody.pubsub.Unsub(s.subChan, topicNames...)
	} else {
		var str = ""
		for _, topicName := range topicNames {
			str += topicName + ","
		}
		str = str[0 : len(str)-1]
		s.melody.errorHandler(s, errors.New("error of unsub current channel,"+str))
	}
}

func (s *Session) writeMessage(message *envelope) {

	defer func() {
		if recover() != nil {
			s.melody.errorHandler(s, ErrWriteToCloseSessionForRecover)
		}
	}()

	if s.closed() {
		s.melody.errorHandler(s, ErrWriteToCloseSession)
		return
	}

	select {
	case s.output <- message:
	default:
		s.melody.errorHandler(s, ErrSessionMessageBufferIsFull)
	}
}

func (s *Session) writeRaw(message *envelope) error {
	if s.closed() {
		return errors.New("tried to write to a closed session")
	}

	s.conn.SetWriteDeadline(time.Now().Add(s.melody.Config.WriteWait))
	err := s.conn.WriteMessage(message.t, message.msg)

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
		if s.open {
			s.conn.Close()
			close(s.closeOutputChan)
			s.melody.pubsub.Unsub(s.subChan)
		}
		s.open = false
		s.rwmutex.Unlock()
	}
}

func (s *Session) ping() {
	s.writeRaw(&envelope{t: websocket.PingMessage, msg: []byte{}})
}

func (s *Session) writePump() {
	ticker := time.NewTicker(s.melody.Config.PingPeriod)
	defer ticker.Stop()

loop:
	for {
		select {
		case _, ok := <-s.closeOutputChan:
			if !ok {
				close(s.output)
				break loop
			}
		case msg, ok := <-s.subChan:
			if !ok {
				break loop
			}
			err := s.writeRaw(msg)
			if err != nil {
				s.melody.errorHandler(s, err)
				break loop
			}

			if msg.t == websocket.CloseMessage {
				break loop
			}

			if msg.t == websocket.TextMessage {
				s.melody.messageSentHandler(s, msg.msg)
			}

			if msg.t == websocket.BinaryMessage {
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

			if msg.t == websocket.CloseMessage {
				break loop
			}

			if msg.t == websocket.TextMessage {
				s.melody.messageSentHandler(s, msg.msg)
			}

			if msg.t == websocket.BinaryMessage {
				s.melody.messageSentHandlerBinary(s, msg.msg)
			}
		case <-ticker.C:
			s.ping()
		}
	}

	log.Println("break loop")
}

func (s *Session) readPump() {
	s.conn.SetReadLimit(s.melody.Config.MaxMessageSize)
	s.conn.SetReadDeadline(time.Now().Add(s.melody.Config.PongWait))

	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(s.melody.Config.PongWait))
		s.melody.pongHandler(s)
		return nil
	})

	if s.melody.closeHandler != nil {
		s.conn.SetCloseHandler(func(code int, text string) error {
			return s.melody.closeHandler(s, code, text)
		})
	}

	for {
		t, message, err := s.conn.ReadMessage()

		if err != nil {
			s.melody.errorHandler(s, err)
			break
		}

		if t == websocket.TextMessage {
			s.melody.messageHandler(s, message)
		}

		if t == websocket.BinaryMessage {
			s.melody.messageHandlerBinary(s, message)
		}
	}
}

// Write writes message to session.
func (s *Session) Write(msg []byte) error {
	if s.closed() {
		return errors.New("session is closed")
	}

	s.writeMessage(&envelope{t: websocket.TextMessage, msg: msg})

	return nil
}

// WriteBinary writes a binary message to session.
func (s *Session) WriteBinary(msg []byte) error {
	if s.closed() {
		return errors.New("session is closed")
	}

	s.writeMessage(&envelope{t: websocket.BinaryMessage, msg: msg})

	return nil
}

// Close closes session.
func (s *Session) Close() error {
	if s.closed() {
		return errors.New("session is already closed")
	}

	s.writeMessage(&envelope{t: websocket.CloseMessage, msg: []byte{}})

	return nil
}

// CloseWithMsg closes the session with the provided payload.
// Use the FormatCloseMessage function to format a proper close message payload.
func (s *Session) CloseWithMsg(msg []byte) error {
	if s.closed() {
		return errors.New("session is already closed")
	}

	s.writeMessage(&envelope{t: websocket.CloseMessage, msg: msg})

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
