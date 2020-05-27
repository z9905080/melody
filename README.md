# melody

[![Build Status](https://travis-ci.org/z9905080/melody.svg)](https://travis-ci.org/z9905080/melody)
[![codecov](https://codecov.io/gh/z9905080/melody/branch/master/graph/badge.svg)](https://codecov.io/gh/z9905080/melody)
[![GoDoc](https://godoc.org/github.com/z9905080/melody?status.svg)](https://godoc.org/github.com/z9905080/melody)

> :notes: Minimalist websocket framework for Go.
> :notes: This is fork project by github.com/olahol/melody

## Install

```bash
go get github.com/z9905080/melody
```

## [Example: DialOption]

Using [Gin](https://github.com/gin-gonic/gin):
```go
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
```


### [More examples](https://github.com/z9905080/melody/tree/master/examples)

## [Documentation](https://godoc.org/github.com/z9905080/melody)

## Contributors

* Shouting (@z9905080)
* Ola Holmström (@olahol)
* Shogo Iwano (@shiwano)
* Matt Caldwell (@mattcaldwell)
* Heikki Uljas (@huljas)
* Robbie Trencheny (@robbiet480)
* yangjinecho (@yangjinecho)

## FAQ

If you are getting a `403` when trying  to connect to your websocket you can [change allow all origin hosts](http://godoc.org/github.com/gorilla/websocket#hdr-Origin_Considerations):

```go
m := melody.New()
m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
```
