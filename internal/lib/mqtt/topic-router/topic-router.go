package topicRouter

import "github.com/eclipse/paho.golang/paho"

type TopicRouter struct {
	*paho.StandardRouter
	handlers map[string]paho.MessageHandler
}

func NewTopicRouter() *TopicRouter {
	return &TopicRouter{
		StandardRouter: paho.NewStandardRouter(),
		handlers:       make(map[string]paho.MessageHandler),
	}
}

func (r *TopicRouter) RegisterHandler(topic string, handler paho.MessageHandler) {
	r.handlers[topic] = handler
	r.StandardRouter.RegisterHandler(topic, handler)
}

func (r *TopicRouter) Topics() []string {
	topics := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		topics = append(topics, t)
	}
	return topics
}
