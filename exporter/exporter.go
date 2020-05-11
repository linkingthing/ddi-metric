package exporter

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ben-han-cn/cement/shell"
	ct "github.com/linkingthing/ddi-metric/collector"
	"github.com/linkingthing/ddi-metric/utils/boltoperation"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*const (
	QuerysPath = "querys"
)*/

type MetricsHandler struct {
	ConfigPath    string
	Ticker        *time.Ticker
	HistoryLength int
	Period        int
	dbHandler     *boltoperation.BoltHandler
	dbPath        string
	Metrics       *ct.Metrics
}

func NewMetricsHandler(path string, length int, period int, dbPath string) *MetricsHandler {
	instance := MetricsHandler{ConfigPath: path, HistoryLength: length, Period: period, dbPath: dbPath}
	instance.Ticker = time.NewTicker(time.Duration(instance.Period) * time.Second)
	instance.dbHandler = boltoperation.NewBoltHandler(instance.dbPath, "dnsmetrics.db")
	instance.Metrics = ct.NewMetrics("dns", instance.dbHandler)
	return &instance
}

func (h *MetricsHandler) Statics() error {
	var err error
	for {
		select {
		case <-h.Ticker.C:
			var para1 string
			var para2 string
			para1 = "-c" + h.ConfigPath + "/rndc.conf"
			para2 = "stats"
			if _, err = shell.Shell(h.ConfigPath+"/rndc", para1, para2); err != nil {
				fmt.Println(err)
			}
			for {
				if _, err := os.Stat(h.ConfigPath + "/named.stats"); err == nil {
					break
				}
			}
			h.QueryStatics()
			h.RecurQueryStatics()
			h.MemHitStatics()
			h.RetCodeStatics("NOERROR", ct.NOERRORPath)
			h.RetCodeStatics("SERVFAIL", ct.SERVFAILPath)
			h.RetCodeStatics("NXDOMAIN", ct.NXDOMAINPath)
			h.RetCodeStatics("REFUSED", ct.REFUSEDPath)
			//remove the named.stats
			if err := os.Remove(h.ConfigPath + "/named.stats"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *MetricsHandler) QueryStatics() error {
	//get the timestamp
	var para1 string
	para1 = "Dump ---"
	var para2 string
	para2 = h.ConfigPath + "/named.stats"
	var value string
	var err error
	if value, err = shell.Shell("grep", para1, para2); err != nil {
		return err
	}
	s := strings.Split(value, "\n")
	var curr []byte
	if len(s) > 1 {
		for _, v := range s[len(s)-2] {
			if v >= '0' && v <= '9' {
				curr = append(curr, byte(v))
			}
		}
	}
	//get the num of query
	var currQuery []byte
	para1 = "QUERY"
	para2 = h.ConfigPath + "/named.stats"
	if value, err = shell.Shell("grep", para1, para2); err != nil {
		return err
	}
	querys := strings.Split(value, "\n")
	if len(querys) > 1 {
		for _, v := range querys[len(querys)-2] {
			if v >= '0' && v <= '9' {
				currQuery = append(currQuery, byte(v))
			}
		}
	}
	h.SaveToDB(string(curr), currQuery, ct.QuerysPath)
	return nil
}

func (h *MetricsHandler) MemHitStatics() error {
	//get the timestamp
	var para1 string
	para1 = "Dump ---"
	var para2 string
	para2 = h.ConfigPath + "/named.stats"
	var value string
	var err error
	if value, err = shell.Shell("grep", para1, para2); err != nil {
		return err
	}
	s := strings.Split(value, "\n")
	var curr []byte
	if len(s) > 1 {
		for _, v := range s[len(s)-2] {
			if v >= '0' && v <= '9' {
				curr = append(curr, byte(v))
			}
		}
	}
	//get the num of recursive query
	para1 = "cache hits (from query)"
	para2 = h.ConfigPath + "/named.stats"
	if value, err = shell.Shell("grep", para1, para2); err != nil {
		return err
	}
	querys := strings.Split(value, "\n")
	var total int
	for _, v := range querys {
		var stringNum []byte
		for _, vv := range v {
			if vv >= '0' && vv <= '9' {
				stringNum = append(stringNum, byte(vv))
			}
		}
		var err error
		var num int
		if num, err = strconv.Atoi(string(stringNum)); err != nil {
			break
		}
		total += num
	}
	h.SaveToDB(string(curr), []byte(strconv.Itoa(total)), ct.MemHitPath)
	return nil
}

func (h *MetricsHandler) RecurQueryStatics() error {
	//get the timestamp
	var para1 string
	para1 = "Dump ---"
	var para2 string
	para2 = h.ConfigPath + "/named.stats"
	var value string
	var err error
	if value, err = shell.Shell("grep", para1, para2); err != nil {
		return err
	}
	s := strings.Split(value, "\n")
	var curr []byte
	if len(s) > 1 {
		for _, v := range s[len(s)-2] {
			if v >= '0' && v <= '9' {
				curr = append(curr, byte(v))
			}
		}
	}
	//get the num of recursive query
	para1 = "queries sent"
	para2 = h.ConfigPath + "/named.stats"
	if value, err = shell.Shell("grep", para1, para2); err != nil {
		return err
	}
	querys := strings.Split(value, "\n")
	var total int
	for _, v := range querys {
		var stringNum []byte
		for _, vv := range v {
			if vv >= '0' && vv <= '9' {
				stringNum = append(stringNum, byte(vv))
			}
		}
		var err error
		var num int
		if num, err = strconv.Atoi(string(stringNum)); err != nil {
			break
		}
		total += num
	}
	h.SaveToDB(string(curr), []byte(strconv.Itoa(total)), ct.RecurQuerysPath)
	return nil
}

func (h *MetricsHandler) RetCodeStatics(retCode string, table string) error {
	//get the timestamp
	var para1 string
	para1 = "Dump ---"
	var para2 string
	para2 = h.ConfigPath + "/named.stats"
	var value string
	var err error
	if value, err = shell.Shell("grep", para1, para2); err != nil {
		return err
	}
	s := strings.Split(value, "\n")
	var curr []byte
	if len(s) > 1 {
		for _, v := range s[len(s)-2] {
			if v >= '0' && v <= '9' {
				curr = append(curr, byte(v))
			}
		}
	}
	//get the num of RetCode
	para1 = "-E"
	para2 = "[0-9]+ " + retCode + "$"
	para3 := h.ConfigPath + "/named.stats"
	if value, err = shell.Shell("grep", para1, para2, para3); err != nil {
		return err
	}
	var currQuery []byte
	querys := strings.Split(value, "\n")
	if len(querys) == 2 {
		for _, v := range querys[0] {
			if v >= '0' && v <= '9' {
				currQuery = append(currQuery, byte(v))
			}
		}
	}
	h.SaveToDB(string(curr), currQuery, table)
	return nil
}

func (h *MetricsHandler) SaveToDB(key string, value []byte, table string) error {
	//check wether there is two pair of values.if it's then delete the old one
	values, err := h.dbHandler.TableKVs(table)
	if err != nil {
		return err
	}
	var timeStamps []string
	for k, _ := range values {
		timeStamps = append(timeStamps, k)
	}
	var delKeys []string
	count := len(values)
	i := 0
	sort.Strings(timeStamps)
	for i < count-h.HistoryLength+1 {
		delKeys = append(delKeys, timeStamps[i])
		i++
	}
	newKVs := map[string][]byte{key: value}
	if err := h.dbHandler.DeleteKVs(table, delKeys); err != nil {
		return err
	}
	if err := h.dbHandler.AddKVs(table, newKVs); err != nil {
		return err
	}
	return nil
}

func (h *MetricsHandler) DNSExporter(port string, metricsPath string, metricsNamespace string) {
	//metrics := ct.NewMetrics(metricsNamespace,h.dbHandler)
	registry := prometheus.NewRegistry()
	registry.MustRegister(h.Metrics)

	http.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>A Prometheus Exporter</title></head>
			<body>
			<h1>A Prometheus Exporter</h1>
			<p><a href='/metrics'>Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Starting Server at http://localhost:%s%s", port, metricsPath)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
