package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"serverlinks/bdrmaplink"
	"serverlinks/config"
	"serverlinks/sptraceroute"
	"serverlinks/utils"
	"strings"
	"time"
)

var Param *config.Config

const PROJECTDIR = "/scratch/cloudspeedtest"
const GETLINKSCRIPT = PROJECTDIR + "/analysis/scripts/get_rtr_links_bdrmap.py"

func main() {
	defer log.Println("End")
	//var wg sync.WaitGroup
	linkmap := make(map[string]*bdrmaplink.Link)
	faripmap := make(map[string][]*bdrmaplink.Link)
	servermap := make(map[string]*sptraceroute.ServerLink)
	if exit := ReadConfig(); exit {
		return
	}
	tmpdir, err := ioutil.TempDir("./", "csp")
	if err != nil {
		log.Panic(err)
	}

	Param.Tmpdir, _ = filepath.Abs(tmpdir)
	log.Println("Tempdir", Param.Tmpdir)
	if Param.Cleanup {
		defer Cleanup(Param.Tmpdir)
	}
	PrepareData()
	//if the link file exist, we do not update the archive
	if Param.AddResult && len(Param.LinkFile) == 0 {
		defer StoreResult()
	}
	prefixtree := utils.PrepareIP2ASTrie(Param.Prefix2ASFile)
	GenerateLinks(linkmap, faripmap)
	sptraceroute.ParseServerTrace(Param, linkmap, faripmap, servermap, prefixtree, sptraceroute.Comcast)
	sptraceroute.ParseServerTrace(Param, linkmap, faripmap, servermap, prefixtree, sptraceroute.Ookla)
	sptraceroute.ParseServerTrace(Param, linkmap, faripmap, servermap, prefixtree, sptraceroute.Ndt)
	matchserver := 0
	for _, s := range servermap {
		matchserver++
		serveras, _, _ := prefixtree.GetByString(s.ServerIP)
		fmt.Printf("%d Server: %s ServerAS: %s, Near: %s Far: %s Far AS %s\n", s.Type, s.ServerIP, serveras, s.Lnk.NearIP, s.Lnk.FarIP, s.Lnk.FarAS)
	}
	coverlink := 0
	for _, l := range linkmap {
		if l.Covered {
			coverlink++
		}
	}
	log.Println("Match Server:", matchserver, " Covered link:", coverlink, "Total Link:", len(faripmap))
	ExportServerLink(servermap)
}

func ReadConfig() bool {
	Param = &config.Config{}
	help := false
	flag.StringVar(&Param.ScamperBin, "scamper", filepath.Join(PROJECTDIR, "tools/scamper/bin/"), "path to scamper binaries")
	flag.StringVar(&Param.ResultFile, "o", "", "path to result file to be analyze (assume in tar.bz format)")
	flag.StringVar(&Param.SiblingDir, "sb", filepath.Join(PROJECTDIR, "analysis/sibling"), "directory that contains sibling files")
	flag.StringVar(&Param.Prefix2ASDir, "pfx", filepath.Join(PROJECTDIR, "analysis/prefix2as"), "directory that contains prefix2as files")
	flag.StringVar(&Param.PeeringDir, "ixp", filepath.Join(PROJECTDIR, "analysis/peering"), "directory that contains peering (ixp) files")
	flag.StringVar(&Param.DelegationDir, "d", filepath.Join(PROJECTDIR, "analysis/delegation/"), "directory that contains delegation files")
	flag.StringVar(&Param.ASRelDir, "a", filepath.Join(PROJECTDIR, "analysis/as-rel/"), "directory that contains AS relationship files")
	flag.StringVar(&Param.SpeedTestServersDir, "st", filepath.Join(PROJECTDIR, "testservers/"), "directory that contains Speedtest servers information")
	flag.BoolVar(&Param.Cleanup, "x", true, "Delete tmp directory after analysis")
	flag.BoolVar(&Param.Clean, "c", false, "Force to regenerate router/link/alias files")
	flag.BoolVar(&Param.AddResult, "A", true, "Add router/link/alias files into original result archive. Note that original file will be replaced")
	flag.BoolVar(&help, "h", false, "Print this help")
	flag.Parse()
	if help {
		fmt.Println("This script extracts interdomain links from bdrmap runs and select speedtest servers.")
		flag.PrintDefaults()
		return help
	}
	if len(Param.ResultFile) == 0 {
		log.Panic("No result file was provided.")
	}
	if _, err := os.Stat(Param.ResultFile); os.IsNotExist(err) {
		log.Panic("Result file does not exist.", Param.ResultFile)
	}
	if _, err := os.Stat(Param.DelegationDir); os.IsNotExist(err) {
		log.Panic("Delegation Directory does not exist.", Param.DelegationDir)
	}
	if _, err := os.Stat(Param.ASRelDir); os.IsNotExist(err) {
		log.Panic("AS Relationship Directory does not exist.", Param.ASRelDir)
	}
	if _, err := os.Stat(Param.SpeedTestServersDir); os.IsNotExist(err) {
		log.Panic("SpeedTest Server Directory does not exist.", Param.SpeedTestServersDir)
	}
	if _, err := os.Stat(Param.SiblingDir); os.IsNotExist(err) {
		log.Panic("Sibling Directory does not exist.", Param.SiblingDir)
	}
	if _, err := os.Stat(Param.Prefix2ASDir); os.IsNotExist(err) {
		log.Panic("Prefix2AS Directory does not exist.", Param.Prefix2ASDir)
	}
	if _, err := os.Stat(Param.PeeringDir); os.IsNotExist(err) {
		log.Panic("Peering Directory does not exist.", Param.PeeringDir)
	}
	log.Println("Result File:", Param.ResultFile)
	//infer other file locations
	return false
}

func PrepareData() int {
	cmd := exec.Command("tar", "xjf", Param.ResultFile, "-C", Param.Tmpdir)
	err := cmd.Run()
	if err != nil {
		log.Panic("Decompress result file failed", err)
	}
	resultdir := filepath.Join(Param.Tmpdir, "results/bdrmap")
	files, err := ioutil.ReadDir(resultdir)
	if err != nil {
		log.Panic(err)
	}
	for _, f := range files {
		if strings.Contains(f.Name(), "bdrmap.meta") {
			Param.MetaFile = filepath.Join(resultdir, f.Name())
			continue
		}
		if strings.Contains(f.Name(), "bdrmap.warts") {
			Param.BdrWartsFile = filepath.Join(resultdir, f.Name())
			continue
		}
		if strings.Contains(f.Name(), "comcast") {
			Param.ComcastWartsFile = filepath.Join(resultdir, f.Name())
			continue
		}
		if strings.Contains(f.Name(), "ndt") {
			Param.NDTWartsFile = filepath.Join(resultdir, f.Name())
			continue
		}
		if strings.Contains(f.Name(), "ookla") {
			Param.OoklaWartsFile = filepath.Join(resultdir, f.Name())
			continue
		}
	}
	if !Param.Clean {
		//Check for router/link/alias files from previous runs
		rootfiles, err := ioutil.ReadDir(Param.Tmpdir)
		if err != nil {
			log.Panic(err)
		}
		for _, f := range rootfiles {
			if strings.Contains(f.Name(), "bdrmap.aliases.out") {
				Param.AliasesFile = filepath.Join(Param.Tmpdir, f.Name())
				continue
			}
			if strings.Contains(f.Name(), "bdrmap.links.out") {
				Param.LinkFile = filepath.Join(Param.Tmpdir, f.Name())
			}
			if strings.Contains(f.Name(), "bdrmap.router.txt") {
				Param.RTRFile = filepath.Join(Param.Tmpdir, f.Name())
			}
		}
	}
	serverfiles, err := ioutil.ReadDir(Param.SpeedTestServersDir)
	if err != nil {
		log.Panic(err)
	}
	//pull the latest server files
	ndtts, ooklats, comcastts := time.Time{}, time.Time{}, time.Time{}
	for _, f := range serverfiles {
		if strings.Contains(f.Name(), "ndt") {
			if f.ModTime().After(ndtts) {
				Param.NDTServerFile = filepath.Join(Param.SpeedTestServersDir, f.Name())
				ndtts = f.ModTime()
				continue
			}
		}
		if strings.Contains(f.Name(), "comcast") {
			if f.ModTime().After(comcastts) {
				Param.ComcastServerFile = filepath.Join(Param.SpeedTestServersDir, f.Name())
				comcastts = f.ModTime()
				continue
			}
		}
		if strings.Contains(f.Name(), "ookla") {
			if f.ModTime().After(ooklats) {
				Param.OoklaServerFile = filepath.Join(Param.SpeedTestServersDir, f.Name())
				ooklats = f.ModTime()
				continue
			}
		}
	}

	mdata, err := ioutil.ReadFile(Param.MetaFile)
	if err != nil {
		log.Panic("Failed to read Meta file:", Param.MetaFile, err)
		return -1
	}
	mdataarr := strings.Split(strings.TrimSuffix(string(mdata), "\n"), ",")
	//expect: <unix timestamp>,<peering file>,<sibling file>,<prefix2as file>
	//example: 1586944209,/home/ubuntu/outdir/datafiles/202001.v4.peering,/home/ubuntu/outdir/datafiles/amazon.sibling.active,/home/ubuntu/outdir/datafiles/20200401.prefix2as
	if len(mdataarr) != 4 {
		log.Panic("Incorrect format in meta file")
		return -1
	}
	Param.PeeringFile = filepath.Join(Param.PeeringDir, filepath.Base(mdataarr[1]))

	if _, err := os.Stat(Param.PeeringFile); os.IsNotExist(err) {
		log.Panic("Peering File does not exist.", Param.PeeringFile)
	}
	//remove active suffix, replace with txt
	tmpsib := filepath.Base(mdataarr[2])
	siblingext := filepath.Ext(mdataarr[2])
	Param.SiblingFile = filepath.Join(Param.SiblingDir, tmpsib[:len(tmpsib)-len(siblingext)]+".txt")
	if _, err := os.Stat(Param.SiblingFile); os.IsNotExist(err) {
		log.Panic("Sibling File does not exist.", Param.SiblingFile)
	}
	ip2as := filepath.Base(mdataarr[3])
	Param.Prefix2ASFile = filepath.Join(Param.Prefix2ASDir, ip2as)
	if _, err := os.Stat(Param.Prefix2ASFile); os.IsNotExist(err) {
		log.Panic("Prefix2AS File does not exist.", Param.Prefix2ASFile, "|")
	}
	//find the AS relationship of the same month as Prefix2AS
	//assume yyyymmdd.prefix2as->yyyymmdd.as-rel.txt
	yrmondd := ip2as[:len(ip2as)-len(filepath.Ext(ip2as))]
	Param.ASRelFile = filepath.Join(Param.ASRelDir, yrmondd+".as-rel.txt")
	if _, err := os.Stat(Param.ASRelFile); os.IsNotExist(err) {
		log.Panic("AS relationship file does not exist.", Param.ASRelFile)
	}
	yrmon := yrmondd[:len(yrmondd)-2]
	Param.DelegationFile = filepath.Join(Param.DelegationDir, "delegated-ipv4-"+yrmon+".txt")
	if _, err := os.Stat(Param.DelegationFile); os.IsNotExist(err) {
		log.Panic("Delegation file does not exist.", Param.DelegationFile)
	}
	return 0
}

func Cleanup(tmpdir string) {
	log.Println("Cleaning up")
	os.RemoveAll(tmpdir)
}

func StoreResult() {
	log.Println("Storing results into result archive. Please wait...")
	storecmd := exec.Command("tar", "cjf", Param.ResultFile, "-C", Param.Tmpdir, filepath.Base(Param.LinkFile), filepath.Base(Param.RTRFile), filepath.Base(Param.AliasesFile), "results")
	err := storecmd.Run()
	if err != nil {
		log.Panic(err)
	}
}

func ExportServerLink(sl map[string]*sptraceroute.ServerLink) {
	//assume result files is .tar.bz2 lenght to cut=8
	slfilepath := Param.ResultFile[:len(Param.ResultFile)-8] + ".gob"
	log.Println("Exporting Gob:", slfilepath)
	slfile, err := os.Create(slfilepath)
	if err != nil {
		log.Panic(err)
	}
	slencoder := gob.NewEncoder(slfile)
	slencoder.Encode(sl)
	slfile.Close()
}

func GenerateLinks(linkmap map[string]*bdrmaplink.Link, faripmap map[string][]*bdrmaplink.Link) {
	log.Println("GenerateRouterFile start")
	if len(Param.LinkFile) == 0 {
		scbdrmapcmd := exec.Command(filepath.Join(Param.ScamperBin, "sc_bdrmap"), "-d", "routers", "-a", Param.Prefix2ASFile, "-g", Param.DelegationFile, "-r", Param.ASRelFile, "-v", Param.SiblingFile, "-x", Param.PeeringFile, Param.BdrWartsFile)
		var out bytes.Buffer
		scbdrmapcmd.Stdout = &out

		err := scbdrmapcmd.Run()
		if err != nil {
			log.Panic(err)
		}
		// convert from /results/bdrmap/xxx.warts to tmpdir/xxx.router.txt
		Param.RTRFile = filepath.Join(Param.Tmpdir, filepath.Base(Param.BdrWartsFile[:len(Param.BdrWartsFile)-len(filepath.Ext(Param.BdrWartsFile))])) + ".router.txt"

		f, err := os.Create(Param.RTRFile)
		if err != nil {
			log.Panic(err)
		}
		_, err = f.Write(out.Bytes())
		if err != nil {
			log.Panic(err)
		}
		f.Sync()
		log.Println("Writing Router file", Param.RTRFile)
		defer f.Close()

		//run the script to extract links and aliases
		getlnkcmd := exec.Command("python", GETLINKSCRIPT, "-w", Param.BdrWartsFile, "-b", Param.RTRFile, "-s", Param.SiblingFile)
		err = getlnkcmd.Run()
		if err != nil {
			log.Panic(err)
		}
		Param.LinkFile = filepath.Join(Param.Tmpdir, filepath.Base(Param.BdrWartsFile[:len(Param.BdrWartsFile)-len(filepath.Ext(Param.BdrWartsFile))])) + ".links.out"
		Param.AliasesFile = filepath.Join(Param.Tmpdir, filepath.Base(Param.BdrWartsFile[:len(Param.BdrWartsFile)-len(filepath.Ext(Param.BdrWartsFile))])) + ".aliases.out"
	} else {
		log.Println("Use previous link file:", Param.LinkFile)
	}
	bdrmaplink.GenerateLinkmap(Param.LinkFile, linkmap, faripmap)
	log.Println("Get link completed")
}
