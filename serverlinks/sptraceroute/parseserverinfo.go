package sptraceroute

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"serverlinks/utils"
)

type OoklaServer struct {
	Url         string `json:"url"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	CountryCode string `json:"cc"`
	Sponsor     string `json:"sponsor"`
	Id          string `json:"id"`
	Https       int    `json:"https_functional"`
	Host        string `json:"host"`
	IPv4        string `json:"ipv4"`
	ASN         string `json:"asn"`
}

type ComcastServer struct {
	IPv4 string `json:"IPv4Address"`
	IPv6 string `json:"IPv6Address"`
	Name string `json:"Sitename"`
	Host string `json:"Fqdn"`
	Id   int    `json:"ServerId"`
}

type NDTServer struct {
	IPs     []string `json:"ip"`
	IPv4    string   `json:"IPv4"`
	IPv6    string   `json:"IPv6"`
	City    string   `json:"city"`
	Url     string   `json:"url"`
	Host    string   `json:"fqdn"`
	Site    string   `json:"site"`
	Country string   `json:"country"`
	Version []string `json:"version"`
	Lat     float64  `json:"latitude"`
	Lon     float64  `json:"longitude"`
}

type ServerEssentials struct {
	Type       Testplatform
	IPv4       string
	ASN        string
	Identifier string
	VP         string
}

func LoadServer(okdatafile, ccdatafile, mldatafile string) map[string]*ServerEssentials {
	okfile, err := ioutil.ReadFile(okdatafile)
	if err != nil {
		log.Panic(err)
	}
	var allokservers []OoklaServer
	err = json.Unmarshal(okfile, &allokservers)
	if err != nil {
		log.Panic("ooklafile error:", err)
	}
	servermap := make(map[string]*ServerEssentials)
	for _, okserver := range allokservers {
		if _, oexist := servermap[okserver.IPv4]; !oexist {
			servermap[okserver.IPv4] = &ServerEssentials{Type: Ookla, IPv4: okserver.IPv4, Identifier: okserver.Name + " - " + okserver.Sponsor, ASN: okserver.ASN}
		}
	}

	ccfile, err := ioutil.ReadFile(ccdatafile)
	var allccservers []ComcastServer
	err = json.Unmarshal(ccfile, &allccservers)
	if err != nil {
		log.Panic("comcast file error:", err)
	}
	for _, ccserver := range allccservers {
		if _, cexist := servermap[ccserver.IPv4]; !cexist {
			servermap[ccserver.IPv4] = &ServerEssentials{Type: Comcast, IPv4: ccserver.IPv4, Identifier: ccserver.Name, ASN: "7922"}
		}
	}
	log.Println("reading mlab:", mldatafile)
	ndtfile, err := ioutil.ReadFile(mldatafile)
	var allndtservers []NDTServer
	err = json.Unmarshal(ndtfile, &allndtservers)
	if err != nil {
		log.Panic("NDT file error:", err)
	}
	for _, nserver := range allndtservers {
		if _, nexist := servermap[nserver.IPv4]; !nexist {
			_, asn := utils.ResolveNet(nserver.Host)
			servermap[nserver.IPv4] = &ServerEssentials{Type: Ndt, IPv4: nserver.IPv4, Identifier: nserver.Host, ASN: asn}
		}
	}
	return servermap
}
