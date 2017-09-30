package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/d2g/dhcp4client"
)

func sendDhcp(mac net.HardwareAddr, s net.IP) {
	var err error

	//Create a connection to use
	//We need to set the connection ports to 1068 and 1067 so we don't need root access
	c, err := dhcp4client.NewInetSock(dhcp4client.SetLocalAddr(net.UDPAddr{IP: net.IPv4(10, 166, 224, 91), Port: 1068}),
		dhcp4client.SetRemoteAddr(net.UDPAddr{IP: s, Port: 67}))
	if err != nil {
		fmt.Printf("Client Connection Generation: %s\n", err)
		return
	}
	defer c.Close()
	fmt.Printf("Connection:%#v\n", c)

	exampleClient, err := dhcp4client.New(dhcp4client.HardwareAddr(mac), dhcp4client.Connection(c))
	if err != nil {
		fmt.Printf("Error:%v\n", err)
		return
	}
	defer exampleClient.Close()
	fmt.Printf("Client:%#v\n", exampleClient)

	success, acknowledgementpacket, err := exampleClient.Request()

	fmt.Printf("Success:%#v\n", success)
	fmt.Printf("Packet:%#v\n", acknowledgementpacket)
}

func main() {
	var server, mac string
	var serverIP net.IP
	flag.StringVar(&server, "server", "255.255.255.255", "specify the server address")
	flag.StringVar(&mac, "mac", "FA-16-3E-1C-E1-09", "specify the client mac address")
	flag.Parse()

	if serverIP = net.ParseIP(server); serverIP == nil {
		fmt.Printf("server address is invalid: %s\n", server)
		return
	}

	m, err := net.ParseMAC(mac)
	if err != nil {
		fmt.Printf("MAC Error:%v\n", err)
		return
	}

	fmt.Printf("sending server to %s\n", serverIP)

	sendDhcp(m, serverIP)
}
