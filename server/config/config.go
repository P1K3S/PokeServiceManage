package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Log         LogConfig         `mapstructure:"log"`
	Jwt         JwtConfig         `mapstructure:"jwt"`
	Auth        AuthConfig        `mapstructure:"auth"`
	Frp         FrpConfig         `mapstructure:"frp"`
	PortRange   PortRangeConfig   `mapstructure:"port_range"`
	SSH         SSHConfig         `mapstructure:"ssh"`
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
	Cors        CorsConfig        `mapstructure:"cors"`
	Websocket   WebsocketConfig   `mapstructure:"websocket"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Dbname       string `mapstructure:"dbname"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type JwtConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type AuthConfig struct {
	RegisterCode  string `mapstructure:"register_code"`
	AdminUsername string `mapstructure:"admin_username"`
	AdminPassword string `mapstructure:"admin_password"`
}

type FrpConfig struct {
	ServerPort int    `mapstructure:"server_port"`
	AuthToken  string `mapstructure:"auth_token"`
}

type PortRangeConfig struct {
	ProtectedPorts []string `mapstructure:"protected_ports"`
	UserPortMin    int      `mapstructure:"user_port_min"`
	UserPortMax    int      `mapstructure:"user_port_max"`
}

type SSHConfig struct {
	DefaultPort     int    `mapstructure:"default_port"`
	DefaultUser     string `mapstructure:"default_user"`
	Timeout         int    `mapstructure:"timeout"`
	TerminalTimeout int    `mapstructure:"terminal_timeout"`
}

type HealthCheckConfig struct {
	Timeout     int  `mapstructure:"timeout"`
	UsePublicIP bool `mapstructure:"use_public_ip"`
}

type CorsConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type WebsocketConfig struct {
	CheckOrigin bool `mapstructure:"check_origin"`
}

var AppConfig *Config

func InitConfig() error {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	AppConfig = &Config{}
	if err := v.Unmarshal(AppConfig); err != nil {
		return err
	}

	return nil
}

func (p *PortRangeConfig) IsProtectedPort(port int) bool {
	for _, item := range p.ProtectedPorts {
		if strings.Contains(item, "-") {
			parts := strings.SplitN(item, "-", 2)
			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil && port >= start && port <= end {
				return true
			}
		} else {
			p, err := strconv.Atoi(item)
			if err == nil && port == p {
				return true
			}
		}
	}
	return false
}

func (c *CorsConfig) AllowAllOrigins() bool {
	for _, o := range c.AllowOrigins {
		if o == "*" {
			return true
		}
	}
	return false
}

func (c *CorsConfig) Origins() []string {
	var result []string
	for _, o := range c.AllowOrigins {
		if o != "*" {
			result = append(result, o)
		}
	}
	return result
}

func (p *PortRangeConfig) UserPortRangeDesc() string {
	return fmt.Sprintf("%d-%d", p.UserPortMin, p.UserPortMax)
}
