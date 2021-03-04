package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)



func DNSServer() {
	configData := readConfig()
	fmt.Println("Retrieving block lists")
	http.DefaultClient = &http.Client{Timeout: time.Second * 2}
	ServerConn, err := net.ListenUDP("udp", &net.UDPAddr{IP:net.ParseIP(configData.ListenAddress),Port:configData.ListenPort,Zone:""})
	ServerConn.SetReadBuffer(8589935000)
	if err != nil{
		panic(err)
	}
	fmt.Println("Started DNS server on " + ServerConn.LocalAddr().String())
	defer ServerConn.Close()
	buf := make([]byte, 512)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		go handleQuery(configData, *addr, buf, n, *ServerConn, err)
	}
}

func readConfig()config{
	var path string
	if len(os.Args) == 2{
		if _, err := os.Stat(os.Args[1]); !os.IsNotExist(err){
			path = os.Args[1]
		}
	}
	if path == "" {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		path = dir + string(os.PathSeparator) + "config.json"
	}
	if _, err := os.Stat(path); os.IsNotExist(err){
		fmt.Println("Could not find config file. Either place a \"config.json\" file in the same directory as the executable. Pass the path to the config file as an argument.\n\n\t" + os.Args[0] + " /path/to/config.json\n\nConfig Example:\n" +
			"{\n  \"listen_address\": \"0.0.0.0\",\n  \"listen_port\": 53,\n  \"dns_servers\": [\n    \"9.9.9.9\",\n    \"149.112.112.112\"\n  ]\n}")
		os.Exit(1)
	}
	fileData, err := ioutil.ReadFile(path)
	if err !=nil{
		panic(err)
	}
	var configData config
	json.Unmarshal(fileData, &configData)

	if len(configData.DNSServers) == 0{
		fmt.Println("No Servers Found")
		os.Exit(1)
	}
	return configData
}


func handleQuery(configData config, addr net.UDPAddr, query []byte, bufLength int, ServerConn net.UDPConn, err error){
	if err != nil {
		fmt.Println(err)
	}
	parsedMessage := dnsmessage.Message{}
	err = parsedMessage.Unpack(query[:bufLength])
	if err != nil {
		fmt.Println("Malformed DNS query\n" + err.Error())
		return
	}
	for _, serverIP := range configData.DNSServers {
		go doLookup("https://"+serverIP+"/dns-query", addr, query[:bufLength], ServerConn)
	}
}

func doLookup(url string, addr net.UDPAddr, query []byte, ServerConn net.UDPConn) {
	response := SendRequest(query, url)
	if response != nil {
		ServerConn.WriteToUDP(response, &addr)
	}
}


func SendRequest(dnsRequest []byte, url string)[]byte{
	resp, err := http.Post(url, "application/dns-message", bytes.NewBuffer(dnsRequest))
	if err == nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return body
	}
	return nil
}



type config struct {
	ListenAddress string   `json:"listen_address"`
	ListenPort    int      `json:"listen_port"`
	DNSServers  []string `json:"dns_servers"`
}
