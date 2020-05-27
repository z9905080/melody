package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/z9905080/melody"
	"log"
	"net/http"
)

// 一連上伺服器先進入大廳頻道
var publicChannel = "public"

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
		messageHandler(m, s, msg)
	})

	m.HandleMessageBinary(func(s *melody.Session, msg []byte) {
		messageHandler(m, s, msg)
	})

	r.Run(":5000")
}

func messageHandler(server *melody.Melody, session *melody.Session, msg []byte) {

	var clientMsg MessageST
	if err := json.Unmarshal(msg, &clientMsg); err != nil {
		log.Println("Err:", err)
	} else {
		if callback, isExist := getServerFuncMap()[clientMsg.OpCode]; isExist {
			callback(server, session, clientMsg)
		} else {
			log.Println("Method Not Found.")
		}
	}
}
