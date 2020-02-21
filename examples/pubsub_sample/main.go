package main

import (
	"github.com/gin-gonic/gin"
	"github.com/z9905080/melody"
	"net/http"
	"time"
)

var textTopic = "topic1"
var binaryTopic = "topic2"

func main() {
	r := gin.Default()

	m := melody.New(
		melody.DialChannelBufferSize(100),
		melody.DialWriteBufferSize(1024))

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		// subscribe topic
		s.AddSub(textTopic)
	})

	m.HandleMessageBinary(func(s *melody.Session, msg []byte) {
		// subscribe topic
		s.AddSub(binaryTopic)
	})

	go func() {
		for {
			// PubMsg is Send TextMessage
			m.PubMsg([]byte("this is a text message."), false, textTopic)

			// PubTextMsg is Send TextMessage (Same to PubMsg)
			m.PubTextMsg([]byte("this is a text message."), false, textTopic)

			// PubBinaryMsg is Send BinaryMessage
			m.PubBinaryMsg([]byte("this is an binary message."), false, binaryTopic)
			time.Sleep(5 * time.Second)
		}
	}()

	r.Run(":5000")
}
