package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnection(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade websocket error: %v", err)
		return
	}
	defer conn.Close()

	log.Println("websocket client connected")

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("read websocket error: %v", err)
			break
		}

		log.Printf("received: %s", message)

		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("write websocket error: %v", err)
			break
		}
	}

	log.Println("websocket client disconnected")
}
