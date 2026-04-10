package emqx

import (
	"fmt"
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

	opts.SetOnConnectHandler(func(c pahomqtt.Client) {
		logger.Log.Info("emqx connected", zap.String("broker", cfg.Broker))
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

func Close() {
	if DefaultClient != nil {
		DefaultClient.client.Disconnect(250)
	}
}
