package main

import (
	"log"
	"path/filepath"
	"serverlinks/config"
	"strconv"
	"sync"
	"time"
)

func main() {

}

func ProcessTraceroute(vmname string, config *config.TrConfig, wg *sync.WaitGroup, workerchan chan int) {
	defer wg.Done()
	//we use "today", this will simple ignore VMs that do not have data in current month
	today := time.Now()
	resultpath := filepath.Join(config.ResultDir, vmname, strconv.Itoa(today.Year()), strconv.Itoa(int(today.Month())))
	trfiles, err := fileutils.SortResultFile(resultpath, sptraceroute.PraseTraceFileTs, 1)
	if err != nil {
		log.Panic(err)
	}
	if len(trfiles) > 0 {
		montrstatus, err := config.MongoClient.QueryDataStatus(vmname)
		var lastts int64 = 0
		if len(montrstatus.TraceFile) > 0 {
			lastts = sptraceroute.ParseTraceFileTs(montrstatus.TraceFile)
		}
		var curidx int
		for curidx = 0; curidx < len(trfiles); curidx++ {
			curts := sptraceroute.ParseTraceFileTs(trfiles[curidx].Name())
			if curts > 0 && curts > lastts {

			}
		}
	}

}
