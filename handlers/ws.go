package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
	"log"
	"net/http"
)

func WSConnectHandler(c *gin.Context) {
	klog.Errorf("@@@ in WSConnectHandler")
	sessionId := c.Param("sessionid")
	session := TerminalSessions.Get(sessionId)
	if session.Id == "" {
		log.Printf("handleTerminalSession: can't find session '%s'", sessionId)
		return
	}
	klog.Errorf("sessionid is '%s'", sessionId)
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
	if conn == nil {
		klog.Errorf("@@@ conn is nil !!!")
	}
	terminalSession.wsConn = conn
	klog.Infof("@@@ wsConn use conn")
	TerminalSessions.Set(sessionId, terminalSession)
	klog.Errorf("@@@ TerminalSession is %v", TerminalSessions.Get(sessionId))
	terminalSession.Bound <- nil
}
