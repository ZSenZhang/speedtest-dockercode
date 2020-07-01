package bdrmaplink

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

//struct for storing interdomain link
type Link struct {
	NearIP  string
	FarIP   string
	FarAS   string
	Covered bool
}

func GenerateLinkmap(LinkFile string, linkmap map[string]*Link, faripmap map[string][]*Link) {
	linkfile, err := os.Open(LinkFile)
	if err != nil {
		log.Panic(err)
	}
	defer linkfile.Close()

	scanner := bufio.NewScanner(linkfile)
	for scanner.Scan() {
		tmplnk := strings.Split(scanner.Text(), "|")
		if len(tmplnk) == 7 {
			lkey := tmplnk[0] + "-" + tmplnk[1]
			if _, exist := linkmap[lkey]; !exist {
				linkmap[lkey] = &Link{NearIP: tmplnk[0], FarIP: tmplnk[1], FarAS: tmplnk[2], Covered: false}
			}
		}
	}
	for _, linkobj := range linkmap {
		if _, exist := faripmap[linkobj.FarIP]; !exist {
			faripmap[linkobj.FarIP] = []*Link{linkobj}
		} else {
			seenas := false
			for _, flink := range faripmap[linkobj.FarIP] {
				if linkobj.FarAS == flink.FarAS {
					seenas = true
					break
				}
			}
			if !seenas {
				faripmap[linkobj.FarIP] = append(faripmap[linkobj.FarIP], linkobj)
			}
			//		log.Println("Has more than one link have the same far ip:", linkkey, linkobj.FarIP, len(faripmap[linkobj.FarIP]))
		}
	}
	for _, farobj := range faripmap {
		if len(farobj) > 1 {
			fmt.Println("FarIP:", farobj[0].FarIP)
			fmt.Println("  - ")
			for _, link := range farobj {
				fmt.Printf(link.FarAS + " ")
			}
		}
	}
	//i love my bun
}
