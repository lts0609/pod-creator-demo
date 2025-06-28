package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func WSConnectHandler(c *gin.Context) {
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
			"error": "WebSocket连接失败",
		})
		return
	}
	terminalSession := session
	terminalSession.wsConn = conn
	TerminalSessions.Set(sessionID, terminalSession)
	terminalSession.Bound <- nil
}
