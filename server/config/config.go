package config

import "github.com/spf13/viper"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
	Frp      FrpConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Dbname       string
	Charset      string
	MaxIdleConns int
	MaxOpenConns int
}

type LogConfig struct {
	Level      string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

type FrpConfig struct {
	ServerPort int    `mapstructure:"server_port"`
	AuthToken  string `mapstructure:"auth_token"`
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
