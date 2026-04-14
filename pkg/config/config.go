package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	EMQX     EMQXConfig     `mapstructure:"emqx"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	LogLevel     string `mapstructure:"log_level"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type KafkaConfig struct {
	Brokers  []string `mapstructure:"brokers"`
	Topic    string   `mapstructure:"topic"`
	GroupID  string   `mapstructure:"group_id"`
	RetryMax int      `mapstructure:"retry_max"`
	Timeout  int      `mapstructure:"timeout"`
}

func (k *KafkaConfig) BrokersString() string {
	return strings.Join(k.Brokers, ",")
}

type EMQXConfig struct {
	Broker     string `mapstructure:"broker"`
	ClientID   string `mapstructure:"client_id"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	QoS        byte   `mapstructure:"qos"`
	KeepAlive  int    `mapstructure:"keepalive"`
	TLSEnabled bool   `mapstructure:"tls_enabled"`
	CACert     string `mapstructure:"ca_cert"`
	Topic      string `mapstructure:"topic"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

var Conf *Config

func Load(path string) error {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file failed: %w", err)
	}

	Conf = &Config{}
	if err := viper.Unmarshal(Conf); err != nil {
		return fmt.Errorf("unmarshal config failed: %w", err)
	}

	return nil
}
