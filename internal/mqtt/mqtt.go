package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	topicRouter "smart-pc-pc-service/internal/lib/mqtt/topic-router"
	pcLogs "smart-pc-pc-service/internal/mqtt/handlers/pc-logs"

	"smart-pc-pc-service/internal/config"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type Connection struct {
	clientConfig autopaho.ClientConfig
	router       *topicRouter.TopicRouter
	log          *slog.Logger
	cm           *autopaho.ConnectionManager
}

func New(log *slog.Logger, cfg config.MQTT) (*Connection, error) {
	const op = "mqtt.New"

	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse mqtt url: %w", op, err)
	}

	router := topicRouter.NewTopicRouter()
	router.RegisterHandler("users/+/pcs/+/log", pcLogs.New(log))

	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     cfg.KeepAlive,
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         cfg.SessionExpiryInterval,
		ReconnectBackoff:              autopaho.NewConstantBackoff(cfg.ReconnectInterval),
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			log.Info("mqtt connection up")
		},
		OnConnectError: func(err error) {
			log.Error("failed to connect to mqtt server", sl.Err(err))
		},
		ClientConfig: paho.ClientConfig{
			ClientID: cfg.ClientID,
			Router:   router,
			OnClientError: func(err error) {
				log.Error("mqtt client error", sl.Err(err))
			},
			OnServerDisconnect: func(d *paho.Disconnect) {
				log.Error("mqtt disconnected by server", slog.Int("reason", int(d.ReasonCode)))
			},
		},
	}

	return &Connection{clientConfig: cliCfg, router: router, log: log}, nil
}

func (c *Connection) Run(ctx context.Context) error {
	const op = "mqtt.Run"

	cm, err := autopaho.NewConnection(ctx, c.clientConfig)
	if err != nil {
		return fmt.Errorf("%s: failed to create connection: %w", op, err)
	}
	c.cm = cm

	if err := cm.AwaitConnection(ctx); err != nil {
		return fmt.Errorf("%s: failed to await connection: %w", op, err)
	}

	subs := make([]paho.SubscribeOptions, 0)
	for _, topic := range c.router.Topics() {
		subs = append(subs, paho.SubscribeOptions{Topic: topic, QoS: 1})
	}
	sa, err := cm.Subscribe(ctx, &paho.Subscribe{Subscriptions: subs})
	if err != nil {
		if sa != nil {
			return fmt.Errorf(
				"%s: failed to subscribe to topics (reason %v): %w",
				op,
				sa.Reasons,
				err,
			)
		}
		return fmt.Errorf("%s: failed to subscribe to topics: %w", op, err)
	}

	<-cm.Done()
	c.log.Info("mqtt connection done")
	return nil
}

func (c *Connection) Done() <-chan struct{} {
	return c.cm.Done()
}
