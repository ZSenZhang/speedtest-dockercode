package utils

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"

	"github.com/zmap/go-iptree/iptree"
)

//given a hostname:port, return the first ipv4 address, and the asn
func ResolveNet(host string) (string, string) {
	addrs := strings.Split(host, ":")
	if len(host) > 0 {
		ips, err := net.LookupIP(addrs[0])
		if err != nil {
			log.Println(err)
			return "", ""
		}
		if len(ips) > 0 {
			//log.Println(addrs[0], ips)
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
			tmpip := ips[foundv4].To4()
			asnq := net.IPv4(tmpip[3], tmpip[2], tmpip[1], tmpip[0]).String() + ".origin.asn.cymru.com"
			outtxt, err := net.LookupTXT(asnq)
			if err != nil {
				return ips[foundv4].String(), ""
			}
			asnstr := strings.Split(outtxt[0], "|")
			return ips[foundv4].String(), strings.TrimSpace(asnstr[0])
		}
	}
	return "", ""
}

//struct for parsing scamper traceroute. do not parse all the fields here. only extract those required ones.
func PrepareIP2ASTrie(prefix2asfile string) *iptree.IPTree {
	t := iptree.New()
	ipasnfile, err := os.Open(prefix2asfile)
	if err != nil {
		log.Panic(err)
	}
	scanner := bufio.NewScanner(ipasnfile)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), "\n")
		if len(line) > 1 {
			//not comment
			if line[0] != '#' {
				data := strings.Fields(line)
				//expect: 115.146.123.131 32      45903
				if len(data) == 3 {
					t.AddByString(data[0]+"/"+data[1], data[2])
				}
			}
		}
	}
	return t

}
