package sptraceroute

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"path/filepath"
	"serverlinks/bdrmaplink"
	"serverlinks/config"
	"sort"
	"strings"

	"github.com/zmap/go-iptree/iptree"
)

const (
	Comcast = 1
	Ookla   = 2
	Ndt     = 3
)

type Testplatform int

type ServerLink struct {
	Type       Testplatform
	ServerIP   string
	Lnk        *bdrmaplink.Link
	Traceroute *SCTraceroute
}
type ServerLinkwithVP struct {
	VP string
	ServerLink
}

//struct for parsing scamper traceroute. do not parse all the fields here. only extract those required ones.
type SCTraceroute struct {
	Type  string `json:"type"`
	DstIP string `json:"dst"`
	DstAS string `json:"-"`
	Hops  []Hop  `json:"hops"`
}
type Hop struct {
	Addr     string  `json:"addr"`
	ProbeTTL int     `json:"probe_ttl"`
	RTT      float64 `json:"rtt"`
	AS       string  `json:"-"`
}

func ParseServerTrace(Param *config.Config, idlink map[string]*bdrmaplink.Link, farlink map[string][]*bdrmaplink.Link, servermap map[string]*ServerLink, prefixip *iptree.IPTree, platform Testplatform) {
	var tracewarts string
	var out bytes.Buffer
	if prefixip == nil {
		log.Panic("IPTree is nil")
	}
	switch platform {
	case Comcast:
		log.Println("Start Parsing Comcast trace")
		tracewarts = Param.ComcastWartsFile
	case Ndt:
		log.Println("Start Parsing Ndt trace")
		tracewarts = Param.NDTWartsFile
	case Ookla:
		log.Println("Start Parsing Ookla trace")
		tracewarts = Param.OoklaWartsFile
	default:
		log.Panic("No such speed test platform")
	}
	tracejsoncmd := exec.Command(filepath.Join(Param.ScamperBin, "sc_warts2json"), tracewarts)
	tracejsoncmd.Stdout = &out
	err := tracejsoncmd.Run()
	if err != nil {
		log.Panic(err) //o. he is so scary, fear, and need a panic button  >.<
	}
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		var tr SCTraceroute
		if len(line) < 2 {
			continue
		}
		err := json.Unmarshal([]byte(line), &tr)
		if err != nil {
			log.Panic(err)
		}
		if tr.Type == "trace" {
			numhop := len(tr.Hops)
			_, sexist := servermap[tr.DstIP]
			//traceroute with less than 2 hops, or it is a duplicate server, simply skip ^.^
			if numhop >= 2 && !sexist {
				if as, found, err := prefixip.GetByString(tr.DstIP); err == nil && found {
					tr.DstAS = as.(string)
				}
				//sort by hop ttl
				sort.Slice(tr.Hops, func(i, j int) bool { return tr.Hops[i].ProbeTTL < tr.Hops[j].ProbeTTL })
				prevIdx := 0
				for h := 1; h < len(tr.Hops); h++ {
					//consecutive hops. hops. and bunnies hops again.
					if tr.Hops[h].ProbeTTL == (tr.Hops[prevIdx].ProbeTTL + 1) {
						//check for near-far @.@
						key := tr.Hops[prevIdx].Addr + "-" + tr.Hops[h].Addr
						if _, lexist := idlink[key]; lexist {
							//found link. !!!! YAY!!!!! ^.^
							servermap[tr.DstIP] = &ServerLink{Type: platform, ServerIP: tr.DstIP, Lnk: idlink[key], Traceroute: &tr}
							idlink[key].Covered = true
							break
						}
					}
					if _, fexist := farlink[tr.Hops[h].Addr]; fexist {
						//only match far address, assign the first one in the slice for now
						if as, found, err := prefixip.GetByString(tr.Hops[h].Addr); err == nil && found {
							tr.Hops[h].AS = as.(string)
						}
						servermap[tr.DstIP] = &ServerLink{Type: platform, ServerIP: tr.DstIP, Lnk: farlink[tr.Hops[h].Addr][0], Traceroute: &tr}
						farlink[tr.Hops[h].Addr][0].Covered = true
					}
				}
			}
		}
	}
}
