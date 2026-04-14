package emqx

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"

	"iot/pkg/config"
	"iot/pkg/logger"
)

type Client struct {
	client pahomqtt.Client
	cfg    config.EMQXConfig
}

var DefaultClient *Client

// NewClient 创建独立的 EMQX 客户端实例
func NewClient(cfg config.EMQXConfig) (*Client, error) {
	opts := pahomqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetKeepAlive(time.Duration(cfg.KeepAlive) * time.Second).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	if cfg.TLSEnabled {
		tlsConfig, err := newTLSConfig(cfg.CACert)
		if err != nil {
			return nil, fmt.Errorf("create tls config failed: %w", err)
		}
		opts.SetTLSConfig(tlsConfig)
	}

	opts.SetOnConnectHandler(func(c pahomqtt.Client) {
		logger.Log.Info("emqx connected", zap.String("broker", cfg.Broker), zap.String("client_id", cfg.ClientID))
	})

	opts.SetConnectionLostHandler(func(c pahomqtt.Client, err error) {
		logger.Log.Error("emqx connection lost", zap.String("client_id", cfg.ClientID), zap.Error(err))
	})

	client := pahomqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("connect emqx failed: %w", token.Error())
	}

	return &Client{
		client: client,
		cfg:    cfg,
	}, nil
}

func Init(cfg config.EMQXConfig) error {
	opts := pahomqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetKeepAlive(time.Duration(cfg.KeepAlive) * time.Second).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	// TLS 配置
	if cfg.TLSEnabled {
		tlsConfig, err := newTLSConfig(cfg.CACert)
		if err != nil {
			return fmt.Errorf("create tls config failed: %w", err)
		}
		opts.SetTLSConfig(tlsConfig)
		logger.Log.Info("emqx tls enabled", zap.String("ca_cert", cfg.CACert))
	}

	opts.SetOnConnectHandler(func(c pahomqtt.Client) {
		logger.Log.Info("emqx connected", zap.String("broker", cfg.Broker), zap.Bool("tls", cfg.TLSEnabled))
	})

	opts.SetConnectionLostHandler(func(c pahomqtt.Client, err error) {
		logger.Log.Error("emqx connection lost", zap.Error(err))
	})

	opts.SetReconnectingHandler(func(c pahomqtt.Client, opts *pahomqtt.ClientOptions) {
		logger.Log.Info("emqx reconnecting...")
	})

	client := pahomqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("connect emqx failed: %w", token.Error())
	}

	DefaultClient = &Client{
		client: client,
		cfg:    cfg,
	}

	return nil
}

// newTLSConfig 根据配置创建 TLS 配置
func newTLSConfig(caCertPath string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	// 加载 CA 证书（如果提供了路径）
	if caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("read ca cert failed: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append ca cert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// Subscribe 订阅主题
func (c *Client) Subscribe(topic string, handler pahomqtt.MessageHandler) error {
	if token := c.client.Subscribe(topic, c.cfg.QoS, handler); token.Wait() && token.Error() != nil {
		return fmt.Errorf("subscribe topic %s failed: %w", topic, token.Error())
	}
	logger.Log.Info("subscribed topic", zap.String("topic", topic))
	return nil
}

// Publish 发布消息
func (c *Client) Publish(topic string, payload interface{}) error {
	if token := c.client.Publish(topic, c.cfg.QoS, false, payload); token.Wait() && token.Error() != nil {
		return fmt.Errorf("publish to topic %s failed: %w", topic, token.Error())
	}
	return nil
}

// PublishBytes 发布二进制消息（Protobuf 场景）
func (c *Client) PublishBytes(topic string, payload []byte) error {
	if token := c.client.Publish(topic, c.cfg.QoS, false, payload); token.Wait() && token.Error() != nil {
		return fmt.Errorf("publish to topic %s failed: %w", topic, token.Error())
	}
	return nil
}

// Close 断开客户端连接
func (c *Client) Close() {
	c.client.Disconnect(250)
}

func Close() {
	if DefaultClient != nil {
		DefaultClient.client.Disconnect(250)
	}
}
