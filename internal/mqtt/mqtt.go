package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	topicRouter "smart-pc-pc-service/internal/lib/mqtt/topic-router"
	pcLogs "smart-pc-pc-service/internal/mqtt/handlers/pc-logs"
	"smart-pc-pc-service/internal/storage/postgres"

	"smart-pc-pc-service/internal/config"

	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type Connection struct {
	cm  *autopaho.ConnectionManager
	log *slog.Logger
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg config.MQTT,
	storage *postgres.Storage,
) (*Connection, error) {
	const op = "mqtt.New"

	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse mqtt url: %w", op, err)
	}

	router := topicRouter.NewTopicRouter()
	router.RegisterHandler("users/+/pcs/+/log", pcLogs.New(ctx, log, storage.PcLogs))

	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     cfg.KeepAlive,
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         cfg.SessionExpiryInterval,
		ReconnectBackoff:              autopaho.NewConstantBackoff(cfg.ReconnectInterval),
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			log.Info("mqtt connection up")

			subs := make([]paho.SubscribeOptions, 0)
			for _, topic := range router.Topics() {
				subs = append(subs, paho.SubscribeOptions{Topic: topic, QoS: 1})
			}
			if _, err := cm.Subscribe(ctx, &paho.Subscribe{Subscriptions: subs}); err != nil {
				log.Error("failed to subscribe to topics", sl.Err(err))
			}
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

	cm, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create connection: %w", op, err)
	}

	return &Connection{cm: cm, log: log}, nil
}

func (c *Connection) Done() <-chan struct{} {
	return c.cm.Done()
}
