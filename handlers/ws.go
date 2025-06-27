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
	log.Printf("!!!! 000")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "WebSocket 连接失败",
		})
		return
	}
	log.Printf("!!!! 111")
	_, buf, err = conn.ReadMessage()
	if err != nil {
		log.Printf("handleTerminalSession: can't read message from session '%s': %v", sessionID, err)
		return
	}
	log.Printf("!!!! 222")
	if err = json.Unmarshal(buf, &msg); err != nil {
		log.Printf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
		return
	}
	log.Printf("!!!! 333")
	if msg.Op != "bind" {
		log.Printf("handleTerminalSession: expected 'bind' message, got: %s", buf)
		return
	}
	log.Printf("!!!! 444")
	terminalSession = session
	log.Printf("!!!! 555")
	terminalSession.wsConn = conn
	TerminalSessions.Set(sessionID, terminalSession)
	terminalSession.Bound <- nil
	log.Printf("!!!! 666")
}
