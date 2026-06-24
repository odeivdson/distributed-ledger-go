package kafka

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewConfig(brokers []string, topic string, groupID string) Config {
	return Config{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	}
}
