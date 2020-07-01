package sptraceroute

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SPGroup struct {
	VPName   string
	SLinkmap map[string]*ServerLink
}

//MAX number of test can run in each hour
var maxtestperhour float64 = 40

func MergeServerLink(gobfiles []string) map[string][]*ServerLinkwithVP {
	servermaps_arr := make([]*SPGroup, 0)
	//decode the gob
	hostregex := regexp.MustCompile(`(\S+)\.\d+\.gob`)
	for _, gfile := range gobfiles {
		log.Println("Opening Gob:", gfile)
		var slinkdata map[string]*ServerLink
		gf, err := os.Open(gfile)
		if err != nil {
			log.Panic(err)
		}
		gdecode := gob.NewDecoder(gf)
		err = gdecode.Decode(&slinkdata)
		if err != nil {
			log.Panic(err)
		}
		hostnamearr := hostregex.FindStringSubmatch(filepath.Base(gfile))
		servermaps_arr = append(servermaps_arr, &SPGroup{VPName: hostnamearr[1], SLinkmap: slinkdata})
		gf.Close()
	}

	//flattern the struct and group all speed test servers from all VPs that traverse the same link
	//map key: FarIP
	linkmap := make(map[string][]*ServerLinkwithVP)
	for _, smap := range servermaps_arr {
		for _, slink := range smap.SLinkmap {
			if _, lexist := linkmap[slink.Lnk.FarIP]; !lexist {
				linkmap[slink.Lnk.FarIP] = []*ServerLinkwithVP{}
			}
			svp := &ServerLinkwithVP{VP: smap.VPName}
			svp.Type = slink.Type
			svp.ServerIP = slink.ServerIP
			svp.Lnk = slink.Lnk
			svp.Traceroute = slink.Traceroute
			linkmap[slink.Lnk.FarIP] = append(linkmap[slink.Lnk.FarIP], svp)
		}
	}
	return linkmap
}

func LoadServerInfo(datadir string) map[string]*ServerEssentials {
	serverfiles := [4]string{}
	files, err := ioutil.ReadDir(datadir)
	if err != nil {
		log.Panic(err)
	}
	sort.Slice(files, func(i, j int) bool {
		//desc order
		return files[i].ModTime().After(files[j].ModTime())
	})
	for _, f := range files {
		if strings.Contains(f.Name(), "comcast") {
			serverfiles[Comcast] = filepath.Join(datadir, f.Name())
			continue
		}
		if strings.Contains(f.Name(), "ookla") {
			serverfiles[Ookla] = filepath.Join(datadir, f.Name())
			continue
		}
		if strings.Contains(f.Name(), "ndt") {
			serverfiles[Ndt] = filepath.Join(datadir, f.Name())
			continue
		}
	}
	return LoadServer(serverfiles[Ookla], serverfiles[Comcast], serverfiles[Ndt])
}

func SelectSPServerforLink(linkmap map[string][]*ServerLinkwithVP, servermap map[string]*ServerEssentials) map[string]*ServerEssentials {
	bestservers := make(map[string]*ServerEssentials)
	//iterate over all links (FarIP)
	for farip, linkservers := range linkmap {
		var currentbest *ServerLinkwithVP = nil
		directpeer := false
		for sidx, server := range linkservers {
			if currentbest == nil {
				if _, sexist := servermap[server.ServerIP]; sexist {
					currentbest = linkservers[sidx]
					if currentbest.Lnk.FarAS == servermap[currentbest.ServerIP].ASN {
						directpeer = true
					}
				}
			} else {
				if _, sexist := servermap[server.ServerIP]; sexist {
					//currentbest is not direct peer, but this one is, replace it.
					if !directpeer && server.Lnk.FarAS == servermap[server.ServerIP].ASN {
						currentbest = linkservers[sidx]
						directpeer = true
					} else if (directpeer && server.Lnk.FarAS == servermap[server.ServerIP].ASN) || (!directpeer && server.Lnk.FarAS != servermap[server.ServerIP].ASN) {
						//compare the RTT to last hop
						if currentbest.Traceroute.Hops[len(currentbest.Traceroute.Hops)-1].RTT > server.Traceroute.Hops[len(server.Traceroute.Hops)-1].RTT {
							currentbest = linkservers[sidx]
						}
					}
				}
			}
		}
		if currentbest != nil {
			spserver := *servermap[currentbest.ServerIP]
			spserver.VP = currentbest.VP
			bestservers[farip] = &spserver
		} else {
			log.Println("Failed to select best server for link:", farip)
		}
	}
	return bestservers
}
func PrintServerList(selectedservers map[string]*ServerEssentials) {
	for farip, ser := range selectedservers {
		fmt.Printf("FarIP: %s, VP: %s, ServerType: %d, ServerIP: %s\n", farip, ser.VP, ser.Type, ser.IPv4)
	}
}

func OutputServerList(selectedservers map[string]*ServerEssentials, outputpath string, splitserver int) {
	var outf *os.File
	if outputpath != "" {
		tmpf, err := os.Create(outputpath)
		if err != nil {
			log.Panic(err)
		}
		outf = tmpf
	} else {
		outf = os.Stdout
	}
	defer outf.Close()
	//0: no spliting, >0 number of measurement rounds in an hour
	if splitserver > 0 {
		serverloadmap := make(map[string][]string)
		for farip, ser := range selectedservers {
			if _, lexist := serverloadmap[ser.VP]; !lexist {
				serverloadmap[ser.VP] = []string{farip}
			} else {
				serverloadmap[ser.VP] = append(serverloadmap[ser.VP], farip)
			}
		}
		rand.Seed(time.Now().UnixNano())
		for sername, farips := range serverloadmap {
			log.Println(sername, len(farips))
			//we assume 1 minute per test. we can run 59 tests in an hour (1 min for data transfer)
			numserver := math.Ceil(float64(len(farips)*splitserver) / maxtestperhour)
			if numserver > 1 {
				//need more than one server
				//assume servername convention: provider-region-number
				sname := strings.Split(sername, "-")
				fmt.Printf("%s-%s: %d target: %d\n", sname[0], sname[1], int(numserver), len(farips))
				rand.Shuffle(len(farips), func(i, j int) { farips[i], farips[j] = farips[j], farips[i] })
				ipperserver := int(math.Floor(float64(len(farips)) / numserver))
				for i := 1; i <= int(numserver); i++ {
					newname := sname[0] + "-" + sname[1] + "-" + strconv.Itoa(i)
					var assignslice []string
					if i < int(numserver) {
						assignslice = farips[(i-1)*ipperserver : i*ipperserver-1]
					} else {
						//last server, all the ip left
						assignslice = farips[(i-1)*ipperserver:]
					}
					for _, farip := range assignslice {
						selectedservers[farip].VP = newname
					}
				}
			}
		}
	}
	for farip, ser := range selectedservers {
		typetext := ""
		switch ser.Type {
		case Comcast:
			typetext = "comcast"
		case Ookla:
			typetext = "ookla"
		case Ndt:
			typetext = "ndt"
		}
		_, err := outf.WriteString(ser.VP + "|" + farip + "|" + ser.ASN + "|" + typetext + "|" + ser.Identifier + "\n")
		if err != nil {
			log.Panic(err)
		}
	}
	outf.Sync()
}
