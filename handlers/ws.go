package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func WSConnectHandler(c *gin.Context) {
	var (
		buf             []byte
		err             error
		msg             TerminalMessage
		terminalSession TerminalSession
	)

	sessionID := c.Param("sessionid")
	session := TerminalSessions.Get(sessionID)
	if session.Id == "" {
		log.Printf("handleTerminalSession: can't find session '%s'", sessionID)
		return
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "WebSocket 连接失败",
		})
		return
	}
	_, buf, err = session.wsConn.ReadMessage()
	if err != nil {
		log.Printf("handleTerminalSession: can't read message from session '%s': %v", sessionID, err)
		return
	}

	if err = json.Unmarshal(buf, &msg); err != nil {
		log.Printf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
		return
	}

	if msg.Op != "bind" {
		log.Printf("handleTerminalSession: expected 'bind' message, got: %s", buf)
		return
	}
	terminalSession.wsConn = conn
	TerminalSessions.Set(sessionID, terminalSession)
	terminalSession.Bound <- nil
}
