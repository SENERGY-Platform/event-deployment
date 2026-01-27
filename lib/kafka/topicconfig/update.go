package topicconfig

import (
	"errors"
	"log/slog"
	"net"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/segmentio/kafka-go"
)

func Ensure(bootstrapUrl string, topic string, config map[string]string) (err error) {
	controller, err := getKafkaController(bootstrapUrl)
	if err != nil {
		slog.Default().Error("unable to find controller", "error", err)
		return err
	}
	if controller == "" {
		slog.Default().Error("unable to find controller", "error", "controller == \"\"")
		return errors.New("unable to find controller")
	}
	return EnsureWithBroker(controller, topic, config)
}

func EnsureWithBroker(broker string, topic string, config map[string]string) (err error) {
	sconfig := sarama.NewConfig()
	sconfig.Version = sarama.V2_4_0_0
	admin, err := sarama.NewClusterAdmin([]string{broker}, sconfig)
	if err != nil {
		return err
	}

	temp := map[string]*string{}
	for key, value := range config {
		tempValue := value
		temp[key] = &tempValue
	}

	err = set(admin, topic, temp)
	if err != nil {
		slog.Default().Warn("unable to update topic --> create with config", "error", err, "topic", topic, "config", config)
		err = create(admin, topic, temp)
	}

	return err
}

func set(admin sarama.ClusterAdmin, topic string, config map[string]*string) (err error) {
	return admin.AlterConfig(sarama.TopicResource, topic, config, false)
}

func create(admin sarama.ClusterAdmin, topic string, config map[string]*string) (err error) {
	return admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
		ConfigEntries:     config,
	}, false)
}

func getKafkaController(bootstrapUrl string) (result string, err error) {
	conn, err := kafka.Dial("tcp", bootstrapUrl)
	if err != nil {
		return result, err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return result, err
	}
	return net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)), nil
}
