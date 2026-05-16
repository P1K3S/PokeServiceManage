package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"service-manage/config"
	"service-manage/model"
	sshutil "service-manage/utils/ssh"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if config.AppConfig.Websocket.CheckOrigin {
			return true
		}
		origin := r.Header.Get("Origin")
		for _, o := range config.AppConfig.Cors.AllowOrigins {
			if o == "*" || o == origin {
				return true
			}
		}
		return false
	},
}

type SSHTerminalHandler struct {
	DB *gorm.DB
}

func NewSSHTerminalHandler(db *gorm.DB) *SSHTerminalHandler {
	return &SSHTerminalHandler{DB: db}
}

func (h *SSHTerminalHandler) HandleTerminal(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var machine model.Machine
	if err := userScope(c, h.DB).First(&machine, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "主机不存在"})
		return
	}

	if !machine.SSHEnabled || machine.SSHUser == "" || machine.SSHPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "该主机未配置SSH连接信息"})
		return
	}

	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	sshClient, err := sshutil.NewClient(&sshutil.Config{
		Host:     machine.IP,
		Port:     machine.SSHPort,
		User:     machine.SSHUser,
		Password: machine.SSHPassword,
		Timeout:  time.Duration(config.AppConfig.SSH.TerminalTimeout) * time.Second,
	})
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\nSSH连接失败: %v\r\n", err)))
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n创建SSH会话失败: %v\r\n", err)))
		return
	}
	defer session.Close()

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return
	}

	session.Stdout = &wsWriter{ws: ws}
	session.Stderr = &wsWriter{ws: ws}

	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "120"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "30"))
	if cols < 10 {
		cols = 120
	}
	if rows < 5 {
		rows = 30
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n请求PTY失败: %v\r\n", err)))
		return
	}

	if err := session.Shell(); err != nil {
		return
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		session.Wait()
	}()

	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				stdinPipe.Close()
				return
			}
			if len(msg) > 0 {
				switch msg[0] {
				case 0:
					stdinPipe.Write(msg[1:])
				case 1:
					if len(msg) >= 3 {
						h := int(msg[1])
						w := int(msg[2])
						if h > 0 && w > 0 {
							session.WindowChange(h, w)
						}
					}
				case 2:
					stdinPipe.Write([]byte{3})
				}
			}
		}
	}()

	select {
	case <-done:
	case <-c.Request.Context().Done():
	}

	uid, uname := getLogUserInfo(c)
	logOperation(h.DB, uid, uname, "connect", "ssh_terminal", uint(id), machine.Name)
}

type wsWriter struct {
	ws *websocket.Conn
}

func (w *wsWriter) Write(p []byte) (int, error) {
	err := w.ws.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
