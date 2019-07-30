package main

import (
	"net/http"
	"shouting/melody"
	"strconv"

	uuid "github.com/satori/go.uuid"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	m := melody.New()

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		uidIntFace, _ := s.Get("uid")
		uid := uidIntFace.(uuid.UUID)
		data := string(msg)
		returnData := []byte("使用者 [" + uid.String() + "] 說：" + data)
		m.Broadcast(returnData)
	})
	m.HandleConnect(func(s *melody.Session) {
		uid := uuid.Must(uuid.NewV4())
		s.Set("uid", uid)
		m.Broadcast([]byte("使用者 [" + uid.String() + "] 已登入！"))
	})

	m.HandleDisconnect(func(s *melody.Session) {
		uidIntFace, _ := s.Get("uid")
		uid := uidIntFace.(uuid.UUID)
		returnData := []byte("使用者 [" + uid.String() + "] 已斷線！ 目前人數：" + strconv.Itoa(m.Len()))
		m.Broadcast(returnData)
	})

	r.Run(":5000")
}
