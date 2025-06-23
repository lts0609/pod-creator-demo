package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"k8s.io/klog/v2"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const END_OF_TRANSMISSION = "\u0004"
const SessionTerminalStoreTime = 5 // session timeout (minute)

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

// TerminalSession implements PtyHandler (using a WebSocket connection)
type TerminalSession struct {
	Id       string
	Bound    chan error
	wsConn   *websocket.Conn
	SizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}
	TimeOut  time.Time
}

type TerminalMessage struct {
	Op, Data, SessionID string
	Rows, Cols          uint16
}

// TerminalSize handles pty->process resize events
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.SizeChan:
		return &size
	case <-t.doneChan:
		return nil
	}
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Read(p []byte) (int, error) {
	klog.Errorf("!!! in Read")
	session := TerminalSessions.Get(t.Id)
	klog.Errorf("!!! line 1")
	if session.TimeOut.Before(time.Now()) {
		klog.Errorf("!!! line 2")
		_ = session.wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(2, "the connection has been disconnected. Please reconnect"))
		return 0, errors.New("the connection has been disconnected. Please reconnect")
	}
	klog.Errorf("!!! line 3")
	TerminalSessions.Set(session.Id, session)
	klog.Errorf("!!! line 4")
	_, message, err := session.wsConn.ReadMessage()
	if err != nil {
		klog.Errorf("!!! line 5")
		// Send terminated signal to process to avoid resource leak
		return copy(p, END_OF_TRANSMISSION), err
	}

	var msg TerminalMessage
	klog.Errorf("!!! line 6")
	if err := json.Unmarshal(message, &msg); err != nil {
		return copy(p, END_OF_TRANSMISSION), err
	}

	switch msg.Op {
	case "stdin":
		klog.Errorf("!!! line 7")
		return copy(p, msg.Data), nil
	case "resize":
		klog.Errorf("!!! line 8")
		session.SizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (t TerminalSession) Write(p []byte) (int, error) {
	klog.Errorf("!!! in Write")
	session := TerminalSessions.Get(t.Id)
	klog.Errorf("!!! line 1")
	if session.TimeOut.Before(time.Now()) {
		klog.Errorf("!!! line 2")
		_ = session.wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(2, "the connection has been disconnected. Please reconnect"))
		return 0, errors.New("the connection has been disconnected. Please reconnect")
	}
	klog.Errorf("!!! line 3")
	TerminalSessions.Set(session.Id, session)
	klog.Errorf("!!! line 4")
	msg, err := json.Marshal(TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}
	klog.Errorf("!!! msg is %v", msg)
	klog.Errorf("!!! line 5")
	if err = session.wsConn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t TerminalSession) Toast(p string) error {
	msg, err := json.Marshal(TerminalMessage{
		Op:   "toast",
		Data: p,
	})
	if err != nil {
		return err
	}

	if err = t.wsConn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return err
	}
	return nil
}

// SessionMap stores a map of all TerminalSession objects and a lock to avoid concurrent conflict
type SessionMap struct {
	Sessions map[string]TerminalSession
	Lock     sync.RWMutex
}

// Get return a given terminalSession by sessionId
func (sm *SessionMap) Get(sessionId string) TerminalSession {
	klog.Errorf("!!! in Get")
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	return sm.Sessions[sessionId]
}

// Set store a TerminalSession to SessionMap
func (sm *SessionMap) Set(sessionId string, session TerminalSession) {
	klog.Errorf("!!! in Set")
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	session.TimeOut = time.Now().Add(SessionTerminalStoreTime * time.Minute)
	sm.Sessions[sessionId] = session
}

// Close shuts down the WebSocket connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (sm *SessionMap) Close(sessionId string, status uint32, reason string) {
	klog.Errorf("!!! in Close")
	if _, ok := sm.Sessions[sessionId]; !ok {
		return
	}
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	err := sm.Sessions[sessionId].wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(int(status), reason))
	if err != nil && status != 1 {
		log.Println(err)
	}

	delete(sm.Sessions, sessionId)
}

// Clean all session when system logout
func (sm *SessionMap) Clean() {
	for _, v := range sm.Sessions {
		v.wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(2, "system is logout, please retry..."))
	}
	sm.Sessions = make(map[string]TerminalSession)
}

var TerminalSessions = SessionMap{Sessions: make(map[string]TerminalSession)}

// 定义 WebSocket 升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleTerminalSession is Called by net/http for any new WebSocket connections
func TerminalSessionHandler(client clientset.Interface, config *rest.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		klog.Errorf("!!! in TerminalSessionHandler")
		podname := c.Query("name")
		namespace := c.Query("namespace")
		if podname == "" || namespace == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "nil podname or namespace parameter",
			})
			return
		}
		shell := c.Query("shell")
		if shell == "" {
			shell = "sh"
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
				"error": "WebSocket failed to connect",
			})
			return
		}
		//defer conn.Close()

		sessionID, err := GenTerminalSessionId()
		if err != nil {
			return
		}
		session := TerminalSession{
			Id:       sessionID,
			wsConn:   conn,
			Bound:    make(chan error, 1),
			SizeChan: make(chan remotecommand.TerminalSize, 10),
			doneChan: make(chan struct{}),
			TimeOut:  time.Now().Add(SessionTerminalStoreTime * time.Minute),
		}
		TerminalSessions.Set(sessionID, session)
		go WaitForTerminal(client, config, namespace, podname, sessionID, shell)
		resp := TerminalResponse{ID: sessionID}
		session.Bound <- nil
		c.Set("terminal", resp)

	}
}

type TerminalResponse struct {
	ID string `json:"id"`
}

func startProcess(k8sClient kubernetes.Interface, cfg *rest.Config, cmd []string, namespace string, podName string, ptyHandler PtyHandler) error {
	klog.Errorf("!!! in startProcess")
	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}
	ctx := context.Background()
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
	if err != nil {
		return err
	}

	return nil
}

func GenTerminalSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}

// isValidShell checks if the shell is an allowed one
func isValidShell(validShells []string, shell string) bool {
	for _, validShell := range validShells {
		if validShell == shell {
			return true
		}
	}
	return false
}

// WaitForTerminal is called from apihandler.handleAttach as a goroutine
// Waits for the WebSocket connection to be opened by the client the session to be Bound in handleTerminalSession
func WaitForTerminal(k8sClient kubernetes.Interface, cfg *rest.Config, namespace string, podName string, sessionId string, shell string) {
	klog.Errorf("!!! in WaitForTerminal")
	select {
	case <-TerminalSessions.Get(sessionId).Bound:
		close(TerminalSessions.Get(sessionId).Bound)
		klog.Errorf("!!! in case1")
		var err error
		validShells := []string{shell}

		if isValidShell(validShells, shell) {
			cmd := []string{shell}
			err = startProcess(k8sClient, cfg, cmd, namespace, podName, TerminalSessions.Get(sessionId))
		} else {
			// No shell given or it was not valid: try some shells until one succeeds or all fail
			// FIXME: if the first shell fails then the first keyboard event is lost
			for _, testShell := range validShells {
				cmd := []string{testShell}
				if err = startProcess(k8sClient, cfg, cmd, namespace, podName, TerminalSessions.Get(sessionId)); err == nil {
					break
				}
			}
		}

		if err != nil {
			TerminalSessions.Close(sessionId, 2, err.Error())
			return
		}

		TerminalSessions.Close(sessionId, 1, "Process exited")
	}
}
