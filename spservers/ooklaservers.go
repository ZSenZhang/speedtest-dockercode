package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
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

type Config struct {
	CreateFile     string
	ReadFile       string
	Workers        int
	ResolveDNS     bool
	SearchASN      string
	SearchProvider string
	SearchCountry  string
	PrintList      string
}

const OoklaQuery = "https://www.speedtest.net/api/js/servers?engine=js&search=%s&https_functional=true&limit=1000"

var cfg *Config
var wg sync.WaitGroup
var allservers []OoklaServer

func main() {
	//	ip, asn := ResolveNet("okc-speedtest.onenet.net.prod.hosts.ooklaserver.net:8080")
	//	fmt.Println(ip, asn)
	cfg = &Config{}
	flag.StringVar(&cfg.ReadFile, "o", "", "Read data from JSON file")
	flag.StringVar(&cfg.CreateFile, "n", "", "Crawl and create JSON file")
	flag.IntVar(&cfg.Workers, "w", 10, "Number of workers")
	flag.StringVar(&cfg.SearchASN, "asn", "", "Print servers in ASN")
	flag.StringVar(&cfg.SearchProvider, "isp", "", "Print servers in ISP")
	flag.StringVar(&cfg.SearchCountry, "country", "", "Print servers in country")
	flag.StringVar(&cfg.PrintList, "p", "", "Print IP of servers to file")
	flag.Parse()

	if len(cfg.ReadFile) > 0 {
		jsonFile, err := os.Open(cfg.ReadFile)
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()
		bvalue, _ := ioutil.ReadAll(jsonFile)
		json.Unmarshal(bvalue, &allservers)
	} else if len(cfg.CreateFile) > 0 {
		allhints := GenerateHints("", 3)
		allservers = make([]OoklaServer, 0)
		rchan := make(chan []OoklaServer)
		go ResultCollector(rchan)
		for i := 1; i <= len(allhints); i++ {
			wg.Add(1)
			go QueryServers(allhints[i-1], rchan)
			if i%cfg.Workers == 0 {
				log.Println("Max worker reached", i, allhints[i])
				wg.Wait()
			}
		}
		wg.Wait()
		close(rchan)
		i := 1
		//resolve all IP and Asn
		for oidx, _ := range allservers {
			wg.Add(1)
			go ResolveIP(oidx)
			if i%cfg.Workers == 0 {
				wg.Wait()
			}
			i++
		}
		wg.Wait()
		file, _ := json.MarshalIndent(allservers, "", " ")
		_ = ioutil.WriteFile(cfg.CreateFile, file, 0644)
		log.Println("Crawled", len(allservers))
	} else {
		fmt.Println("Please provide input/output file name")
	}
	if len(cfg.SearchASN) > 0 && len(cfg.SearchProvider) > 0 && len(cfg.SearchCountry) > 0 {
		log.Fatal("Only can search with one criteria")
	} else if len(cfg.SearchASN) > 0 {
		for _, j := range allservers {
			if j.ASN == cfg.SearchASN {
				fmt.Println(j.IPv4)
			}
		}
	} else if len(cfg.SearchProvider) > 0 {
		for _, j := range allservers {
			if strings.Contains(j.Sponsor, cfg.SearchProvider) {
				fmt.Println(j.IPv4)
			}
		}
	} else if len(cfg.SearchCountry) > 0 {
		for _, j := range allservers {
			if strings.Contains(j.CountryCode, cfg.SearchCountry) {
				fmt.Println(j.IPv4)
			}
		}
	}
	/*	for _, j := range allservers {
		fmt.Println(j.Id, j.Host)
	}*/
}

func QueryServers(hints string, results chan []OoklaServer) {
	qstring := fmt.Sprintf(OoklaQuery, hints)
	response, err := http.Get(qstring)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	rspobj := []OoklaServer{}
	json.Unmarshal(responseData, &rspobj)
	results <- rspobj
	wg.Done()
}

func ResultCollector(rchan chan []OoklaServer) {
	servermap := make(map[string]bool)
	for r := range rchan {
		log.Println("Result contains ", len(r), "records")
		for _, server := range r {
			if _, sexist := servermap[server.Id]; !sexist {
				//only insert those had not seen before
				servermap[server.Id] = true
				allservers = append(allservers, server)
			}
		}
	}

}

//generate all combination of a-z with length len
func GenerateHints(hint string, len int) []string {
	if len <= 0 {
		return []string{hint}
	}
	hints := []string{}
	for a := int('a'); a <= int('z'); a++ {
		newhint := hint + string(a)
		hints = append(hints, GenerateHints(newhint, len-1)...)
	}
	return hints
}

func ResolveIP(serveridx int) {
	allservers[serveridx].IPv4, allservers[serveridx].ASN = ResolveNet(allservers[serveridx].Host)
	wg.Done()
}

//given a hostname:port, return the first ipv4 address, and the asn
func ResolveNet(host string) (string, string) {
	addrs := strings.Split(host, ":")
	if len(host) > 0 {
		ips, err := net.LookupIP(addrs[0])
		if err != nil {
			return "", ""
		}
		if len(ips) > 0 {
			log.Println(addrs[0], ips)
			i := 0
			foundv4 := -1
			for i = 0; i < len(ips); i++ {
				if ips[i].To4() != nil {
					foundv4 = i
					break
				}
			}
			if foundv4 < 0 {
				log.Println("No IPv4 found")
				return "", ""
			}
			asnq := net.IPv4(ips[foundv4][3], ips[foundv4][2], ips[foundv4][1], ips[foundv4][0]).String() + ".origin.asn.cymru.com"
			outtxt, err := net.LookupTXT(asnq)
			if err != nil {
				return ips[foundv4].String(), ""
			}
			//fmt.Println(asnq, outtxt[0])
			asnstr := strings.Split(outtxt[0], "|")
			return ips[foundv4].String(), strings.TrimSpace(asnstr[0])
		}
	}
	return "", ""
}
