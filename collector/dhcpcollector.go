// refer to dns collector, dhcp statistics init here

package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/linkingthing/ddi/dhcp"

	"github.com/linkingthing/ddi/utils"
)

// unmarshall data from kea statistics commands
type CurlKeaArguments struct {
	Pkt4Received []interface{} `json:"pkt4-received"`
}
type CurlKeaStats struct {
	Arguments CurlKeaArguments `json:"arguments"`
	Result    string           `json:"result"`
}

type CurlKeaStatsAll struct {
	Arguments map[string][]interface{} `json:"arguments"`
	Result    string                   `json:"result"`
}

// dashboard -- dhcp -- packet statistics
func (c *Metrics) GenerateDhcpPacketStatistics() error {
	//log.Println("+++ into GenerateDhcpPacketStatistics()")

	//get packet statistics data, export it to prometheus
	//todo move ip:port into conf
	//log.Println("in GenerateDhcpPacketStatistics, keaServer: " + utils.KeaServer)
	url := "http://" + utils.DhcpHost + ":" + dhcp.DhcpPort
	curlCmd := "curl -X POST \"" + url + "\"" + " -H 'Content-Type: application/json' -d '" +
		`   {
	            "command": "statistic-get",
                "service": ["dhcp4"],
	            "arguments": {
	                "name": "pkt4-received"
	            }
	        }
	        ' 2>/dev/null`
	//log.Println("--- GenerateDhcpPacketStatistics curlCmd: ", curlCmd)
	out, err := utils.Cmd(curlCmd)

	if err != nil {
		log.Println("curl error: ", err)
		return err
	}
	//log.Println("+++ GenerateDhcpPacketStatistics(), out")
	//log.Println(out)
	//log.Println("--- GenerateDhcpPacketStatistics(), out")

	var curlRet CurlKeaStats
	json.Unmarshal([]byte(out[1:len(out)-1]), &curlRet)
	maps := curlRet.Arguments.Pkt4Received
	c.gaugeMetricData["dhcppacket"] = float64(len(maps))

	return nil
}

func GetKeaStatisticsAll() *CurlKeaStatsAll {
	//todo move ip:port into conf
	//log.Println("in GetKeaStatisticsAll, keaServer: " + utils.KeaServer)
	url := "http://" + utils.DhcpHost + ":" + dhcp.DhcpPort
	curlCmd := "curl -X POST \"" + url + "\"" + " -H 'Content-Type: application/json' -d '" +
		`   { "command": "statistic-get-all", "service": ["dhcp4"], "arguments": { }}' 2>/dev/null`
	//log.Println("--- GetKeaStatisticsAll curlCmd: ", curlCmd)
	out, err := utils.Cmd(curlCmd)
	if err != nil {
		log.Println("curl error: ", err)
		return nil
	}
	var curlRet CurlKeaStatsAll
	json.Unmarshal([]byte(out[1:len(out)-1]), &curlRet)
	return &curlRet
}

// dashboard -- dhcp -- packet statistics
func (c *Metrics) GenerateDhcpLeasesStatistics() error {
	//log.Println("+++ into GenerateDhcpPacketStatistics()")

	//get packet statistics data, export it to prometheus
	curlRet := GetKeaStatisticsAll()
	if curlRet == nil {
		c.gaugeMetricData["dhcplease"] = float64(0)
		return fmt.Errorf("Get Kea Statistics All error")
	}
	leaseNum := 0
	maps := curlRet.Arguments
	for k, v := range maps {

		//log.Println("in lease statistics(), for loop, k: ", k, ", v: ", v)
		rex := regexp.MustCompile(`^subnet\[(\d+)\]\.assigned-addresses`)
		out := rex.FindAllStringSubmatch(k, -1)
		if len(out) > 0 {
			//log.Println("+++ out: ", out)
			for range out {
				//idx := i[1]
				leaseNum += len(v)
				//log.Println("+++ i: ", i[1], ", len[v], ", len(v), ", leaseNum: ", leaseNum)
			}
		}
	}

	//maps := curlRet.Arguments.Pkt4Received
	c.gaugeMetricData["dhcplease"] = float64(leaseNum)

	return nil
}

// dashboard -- dhcp -- usage statistics
func (c *Metrics) GenerateDhcpUsageStatistics() error {
	//log.Println("+++ into GenerateDhcpPacketStatistics()")

	//get packet statistics data, export it to prometheus
	//todo move ip:port into conf
	//log.Println("in GenerateDhcpUsageStatistics, keaServer: " + utils.KeaServer)
	url := "http://" + utils.DhcpHost + ":" + dhcp.DhcpPort
	curlCmd := "curl -X POST \"" + url + "\"" + " -H 'Content-Type: application/json' -d '" +
		`   {
                "command": "statistic-get-all",
                "service": ["dhcp4"],
                "arguments": { }
	        }
	        ' 2>/dev/null`
	//log.Println("--- GenerateDhcpPacketStatistics curlCmd: ", curlCmd)
	out, err := utils.Cmd(curlCmd)
	if err != nil {
		log.Println("curl error: ", err)
		return err
	}

	var curlRet CurlKeaStatsAll
	leaseNum := 0
	totalNum := 0
	json.Unmarshal([]byte(out[1:len(out)-1]), &curlRet)

	maps := curlRet.Arguments
	for k, v := range maps {
		//log.Println("in lease statistics(), for loop, k: ", k)
		rex := regexp.MustCompile(`^subnet\[(\d+)\]\.(\S+)`)
		out := rex.FindAllStringSubmatch(k, -1)

		if len(out) > 0 {
			for _, i := range out {

				addrType := i[2]
				if addrType == "total-addresses" {
					total := maps[k][0].([]interface{})[0]
					totalNum += int(Decimal(total.(float64)))

				} else if addrType == "assigned-addresses" {
					leaseNum += len(v)
				}
				//log.Println("+++ i: ", i[1], ", len[v], ", len(v), ", leaseNum: ", leaseNum)
			}
		}
	}
	//log.Println("leaseNum: ", leaseNum)
	//log.Println("totalNum: ", totalNum)
	dhcpUsage := 0.0
	if totalNum > 0 {
		dhcpUsage = Decimal(float64(leaseNum) / float64(totalNum) * 100)
	}
	//log.Println("dhcpUsage: ", dhcpUsage)

	//maps := curlRet.Arguments.Pkt4Received
	c.gaugeMetricData["dhcpusage"] = Decimal(float64(dhcpUsage))

	return nil
}

type ValueIntf [3]interface{}

func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}
