package arp

import (
	"fmt"
	"github.com/mdlayher/arp"
	"log"
	"net"
	"time"
)

func Arp(interf string, dstIP string, timeout int) (*net.HardwareAddr, error) {
	ifi, err := net.InterfaceByName(interf)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var ip net.IP
	ip = net.ParseIP(dstIP)
	if ip == nil {
		return nil, fmt.Errorf("ip:%s format error!", dstIP)
	}
	client, err := arp.Dial(ifi)
	if err != nil {
		log.Println("couldn't create ARP client: %s", err)
		return nil, err
	}
	defer client.Close()
	// Set request deadline
	if err := client.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second)); err != nil {
		log.Println(err)
		return nil, err
	}

	var hwAddr net.HardwareAddr
	hwAddr, err = client.Resolve(ip)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &hwAddr, nil
}
