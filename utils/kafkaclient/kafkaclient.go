package kafkaclient

import (
	"context"

	"log"

	kg "github.com/segmentio/kafka-go"
)

type KafkaCliHandler struct {
	dnsTopic          string
	dhcpv4Topic       string
	dhcpv6Topic       string
	kafkaWriterDhcpv4 *kg.Writer
	kafkaWriterDhcpv6 *kg.Writer
	kafkaWriter       *kg.Writer
}

var KafkaClient *KafkaCliHandler

func NewKafkaCliHandler(dnsTopic string, dhcpv4Topic string, dhcpv6Topic string, kafkaServer string) *KafkaCliHandler {
	instance := KafkaCliHandler{}
	instance.dnsTopic = dnsTopic
	instance.dhcpv4Topic = dhcpv4Topic
	instance.dhcpv6Topic = dhcpv6Topic
	instance.kafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{kafkaServer},
		Topic:   dnsTopic,
	})
	instance.kafkaWriterDhcpv4 = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{kafkaServer},
		Topic:   dhcpv4Topic,
	})
	instance.kafkaWriterDhcpv6 = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{kafkaServer},
		Topic:   dhcpv6Topic,
	})
	return &instance
}
func (handler *KafkaCliHandler) SendCmd(data []byte, cmd string) error {
	postData := kg.Message{
		Key:   []byte(cmd),
		Value: data,
	}
	if err := handler.kafkaWriter.WriteMessages(context.Background(), postData); err != nil {
		return err
	}
	return nil
}

func (handler *KafkaCliHandler) SendCmdDhcpv4(data []byte, cmd string) error {
	log.Println("into SendCmdDhcpv4, cmd: ", cmd)
	postData := kg.Message{
		Key:   []byte(cmd),
		Value: data,
	}
	if err := handler.kafkaWriterDhcpv4.WriteMessages(context.Background(), postData); err != nil {
		return err
	}
	return nil
}
func (handler *KafkaCliHandler) SendCmdDhcpv6(data []byte, cmd string) error {
	postData := kg.Message{
		Key:   []byte(cmd),
		Value: data,
	}
	if err := handler.kafkaWriterDhcpv6.WriteMessages(context.Background(), postData); err != nil {
		return err
	}
	return nil
}
