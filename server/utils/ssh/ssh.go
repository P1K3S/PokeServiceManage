package ssh

import (
	"net"
	"strconv"
	"time"

	"service-manage/config"

	"golang.org/x/crypto/ssh"
)

func getDefaultTimeout() time.Duration {
	t := config.AppConfig.SSH.Timeout
	if t <= 0 {
		t = 5
	}
	return time.Duration(t) * time.Second
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  time.Duration
}

func NewClient(cfg *Config) (*ssh.Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = getDefaultTimeout()
	}

	config := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         cfg.Timeout,
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	return ssh.Dial("tcp", addr, config)
}

func RunCommand(cfg *Config, cmd string) (string, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	return string(output), err
}

func CheckConnection(cfg *Config) error {
	if cfg.User != "" && cfg.Password != "" {
		client, err := NewClient(cfg)
		if err != nil {
			return err
		}
		return client.Close()
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = getDefaultTimeout()
	}
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	dialer := net.Dialer{Timeout: cfg.Timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
	return conn.Close()
}
