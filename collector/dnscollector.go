/**
 *  Author: SongLee24
 *  Email: lisong.shine@qq.com
 *  Date: 2018-08-15
 *
 *
 *  prometheus.Desc是指标的描述符，用于实现对指标的管理
 *
 */

package collector

import (
	"github.com/linkingthing/ddi-metric/utils/boltoperation"
	"github.com/prometheus/client_golang/prometheus"
	"sort"
	"strconv"
	"sync"
)

const (
	QuerysPath      = "querys"
	RecurQuerysPath = "recurquerys"
	MemHitPath      = "memhit"
	NOERRORPath     = "NOERROR"
	SERVFAILPath    = "SERVFAIL"
	NXDOMAINPath    = "NXDOMAIN"
	REFUSEDPath     = "REFUSED"
)

// 指标结构体
type Metrics struct {
	metrics           map[string]*prometheus.Desc
	counterMetricData map[string]float64
	gaugeMetricData   map[string]float64
	dbHandler         *boltoperation.BoltHandler
	mutex             sync.Mutex
}

/**
 * 函数：newGlobalMetric
 * 功能：创建指标描述符
 */
func newGlobalMetric(namespace string, metricName string, docString string, labels []string) *prometheus.Desc {
	instance := prometheus.NewDesc(namespace+"_"+metricName, docString, labels, nil)
	return instance
}

/**
 * 工厂方法：NewMetrics
 * 功能：初始化指标信息，即Metrics结构体
 */
func NewMetrics(namespace string, h *boltoperation.BoltHandler) *Metrics {
	instance := &Metrics{
		metrics: map[string]*prometheus.Desc{
			"counter": newGlobalMetric(namespace, "counter", "The description of qps, collectint every minute.", []string{"data_type"}),
			"gauge":   newGlobalMetric(namespace, "gauge", "The description of my_gauge_metric", []string{"data_type"}),
		},
	}
	instance.counterMetricData = make(map[string]float64, 10)
	instance.gaugeMetricData = make(map[string]float64, 10)
	instance.dbHandler = h
	return instance
}

/**
 * 接口：Describe
 * 功能：传递结构体中的指标描述符到channel
 */
func (c *Metrics) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m
	}
}

/**
 * 接口：Collect
 * 功能：抓取最新的数据，传递给channel
 */
func (c *Metrics) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // 加锁
	defer c.mutex.Unlock()
	c.GenerateQPS()
	c.GenerateQuery()
	c.GenerateRecurQuery()
	c.GenerateMemHit()
	c.GenerateRetCode("NOERROR", NOERRORPath)
	c.GenerateRetCode("SERVFAIL", SERVFAILPath)
	c.GenerateRetCode("NXDOMAIN", NXDOMAINPath)
	c.GenerateRetCode("REFUSED", REFUSEDPath)
	// add dhcp statistics here
	c.GenerateDhcpPacketStatistics()
	c.GenerateDhcpLeasesStatistics()
	c.GenerateDhcpUsageStatistics()

	for host, currentValue := range c.counterMetricData {
		ch <- prometheus.MustNewConstMetric(c.metrics["counter"], prometheus.CounterValue, float64(currentValue), host)
	}
	for host, currentValue := range c.gaugeMetricData {
		ch <- prometheus.MustNewConstMetric(c.metrics["gauge"], prometheus.GaugeValue, float64(currentValue), host)
	}
}

/**
 * 函数：GenerateData
 * 功能：生成模拟数据
 */
/*func (c *Metrics) GenerateData(dbHandler *boltoperation.BoltHandler) {
	counterMetricData = map[string]int{
		"yahoo.com":  int(rand.Int31n(1000)),
		"google.com": int(rand.Int31n(1000)),
	}
	gaugeMetricData = map[string]int{
		"yahoo.com":  int(rand.Int31n(10)),
		"google.com": int(rand.Int31n(10)),
	}
	return
}*/

func (c *Metrics) GenerateQPS() error {
	kvs, err := c.dbHandler.TableKVs(QuerysPath)
	if err != nil {
		return err
	}
	var timeStamps []string
	for k, _ := range kvs {
		timeStamps = append(timeStamps, k)
	}
	sort.Strings(timeStamps)
	if len(kvs) > 1 {
		var numPrev int
		if numPrev, err = strconv.Atoi(timeStamps[len(timeStamps)-2]); err != nil {
			return err
		}
		var numLast int
		if numLast, err = strconv.Atoi(timeStamps[len(timeStamps)-1]); err != nil {
			return err
		}
		var queryPrev int
		if queryPrev, err = strconv.Atoi(string(kvs[timeStamps[len(timeStamps)-2]])); err != nil {
			return err
		}
		var queryLast int
		if queryLast, err = strconv.Atoi(string(kvs[timeStamps[len(timeStamps)-1]])); err != nil {
			return err
		}
		c.gaugeMetricData["qps"] = float64(queryLast-queryPrev) / float64(numLast-numPrev)
	}
	return nil
}

func (c *Metrics) GenerateQuery() error {
	kvs, err := c.dbHandler.TableKVs(QuerysPath)
	if err != nil {
		return err
	}
	var timeStamps []string
	for k, _ := range kvs {
		timeStamps = append(timeStamps, k)
	}
	sort.Strings(timeStamps)
	if len(kvs) > 1 {
		var query int
		if query, err = strconv.Atoi(string(kvs[timeStamps[len(timeStamps)-1]])); err != nil {
			return err
		}
		c.counterMetricData["querys"] = float64(query)
	}
	return nil
}

func (c *Metrics) GenerateRecurQuery() error {
	kvs, err := c.dbHandler.TableKVs(RecurQuerysPath)
	if err != nil {
		return err
	}
	var timeStamps []string
	for k, _ := range kvs {
		timeStamps = append(timeStamps, k)
	}
	sort.Strings(timeStamps)
	if len(kvs) > 1 {
		var query int
		if query, err = strconv.Atoi(string(kvs[timeStamps[len(timeStamps)-1]])); err != nil {
			return err
		}
		c.counterMetricData["recurquerys"] = float64(query)
	}
	return nil
}

func (c *Metrics) GenerateMemHit() error {
	kvs, err := c.dbHandler.TableKVs(MemHitPath)
	if err != nil {
		return err
	}
	var timeStamps []string
	for k, _ := range kvs {
		timeStamps = append(timeStamps, k)
	}
	sort.Strings(timeStamps)
	if len(kvs) > 1 {
		var query int
		if query, err = strconv.Atoi(string(kvs[timeStamps[len(timeStamps)-1]])); err != nil {
			return err
		}
		c.counterMetricData["memhit"] = float64(query)
	}
	return nil
}

func (c *Metrics) GenerateRetCode(retCode string, table string) error {
	kvs, err := c.dbHandler.TableKVs(table)
	if err != nil {
		return err
	}
	var timeStamps []string
	for k, _ := range kvs {
		timeStamps = append(timeStamps, k)
	}
	sort.Strings(timeStamps)
	if len(kvs) > 1 {
		var query int
		if query, err = strconv.Atoi(string(kvs[timeStamps[len(timeStamps)-1]])); err != nil {
			return err
		}
		c.counterMetricData[retCode] = float64(query)
	} else if len(kvs) <= 1 {
		c.counterMetricData[retCode] = 0
	}
	return nil
}
