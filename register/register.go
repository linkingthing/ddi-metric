// test for node registeration on boot time
package register

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	KafkaTopicProm = "prom"
	key            = "prom"
	ControllerRole = "controller"
	DNSRole        = "dns"
	DHCPRole       = "dhcp"
	online         = 1
	offline        = 0
)

type PromRole struct {
	Hostname string `json:"hostname"`
	PromHost string `json:"promHost"`
	PromPort string `json:"promPort"`
	IP       string `json:"ip"`
	Role     string `json:"role"`     // 3 roles: Controller, Db, Kafka
	State    uint   `json:"state"`    // 1 online 0 offline
	HbTime   int64  `json:"hbTime"`   //timestamp of most recent heartbeat time
	OnTime   int64  `json:"onTime"`   //timestamp of the nearest online time
	ParentIP string `json:"parentIP"` //parent node ip in node management graph
}

var OnlinePromHosts = make(map[string]PromRole)
var OfflinePromHosts = make(map[string]PromRole)
var KafkaOffset int64 = 0
var KafkaOffsetFile = "/tmp/kafka-offset.txt" // store kafka offset num into this file

var (
	Dhcpv4Topic = "dhcpv4"
	Dhcpv6Topic = "dhcpv6"
)
var KafkaOffsetFileDhcpv4 = "/tmp/kafka-offset-dhcpv4.txt" // store kafka offset num into this file
var KafkaOffsetDhcpv4 int64 = 0

// produceProm node uses kafka to report it's alive state
func ProduceProm(msg kafka.Message, kafkaServerAddr string) {
	log.Println("in utils/kafka-common, KafkaServerProm: ", kafkaServerAddr)
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaServerAddr},
		Topic:   KafkaTopicProm,
	})

	w.WriteMessages(context.Background(), msg)
}

// consumerProm server get msg from kafka topic regularly, if not accept, turn the machine's state to offline
func ConsumerProm(kafkaServerAddr string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{kafkaServerAddr},
		Topic:       KafkaTopicProm,
		StartOffset: kafka.LastOffset,
	})
	defer r.Close()

	//read kafkaoffset from KafkaOffsetFile and set it to KafkaOffset
	size, err := ioutil.ReadFile(KafkaOffsetFile)
	if err == nil {
		offset, err2 := strconv.Atoi(string(size))
		if err2 != nil {
			log.Println(err2)
		}
		KafkaOffset = int64(offset)
		r.SetOffset(KafkaOffset)
	}
	log.Println("kafka Offset: ", KafkaOffset)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("ConsumerProm:", err)
			break
		}
		log.Printf(", message at offset %d: key: %s, value: %s\n", m.Offset, string(m.Key),
			string(m.Value))

		if string(m.Key) == "prom" {
			var Role PromRole
			err := json.Unmarshal(m.Value, &Role)
			if err != nil {
				log.Println(err)
				return
			}
			//put Role struct into OnlinePromHosts map
			Role.OnTime = time.Now().Unix()
			Role.IP = strings.TrimSpace(Role.IP)
			Role.Role = strings.TrimSpace(Role.Role)
			OnlinePromHosts[Role.IP+"_"+Role.Role] = Role

			log.Println("+++ OnlinePromHosts")
			log.Println(OnlinePromHosts)
			log.Println("--- OnlinePromHosts")
		}

		//store curOffset into KafkaOffsetFile
		curOffset := r.Stats().Offset
		if curOffset > KafkaOffset {
			KafkaOffset = curOffset
			byteOffset := []byte(strconv.Itoa(int(curOffset)))
			err = ioutil.WriteFile(KafkaOffsetFile, byteOffset, 0644)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func RegisterNode(hostName string, promHostIP string, promHostPort string, hostIP string, parentIP string, role string, kafkaServerAddr string) error {
	//send kafka msg to topic prom
	var PromInfo = PromRole{
		Hostname: hostName,
		PromHost: promHostIP,
		PromPort: promHostPort,
		IP:       hostIP,
		Role:     role,
		State:    online,
		OnTime:   time.Now().Unix(),
		ParentIP: parentIP,
	}
	PromJson, err := json.Marshal(PromInfo)
	if err != nil {
		fmt.Println(err)
		return err
	}

	value := PromJson
	msg := kafka.Message{
		Topic: KafkaTopicProm,
		Key:   []byte(key),
		Value: []byte(value),
	}
	ProduceProm(msg, kafkaServerAddr)
	return nil
}
