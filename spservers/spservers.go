package main

import (
	"log"
	"spservers/comcast"
	"spservers/common"
	"spservers/spdb"
	"time"
)

func main() {
	cfg := common.Config{StartTime: time.Now(), Workers: 10}
	db := spdb.NewMongoDB("beamermongosp.json", "speedtest")
	//	ookla.LoadOokla(&cfg)
	db.ResetEnable("comcast")
	cser := comcast.LoadComcastServer(&cfg)
	if nser, err := db.InsertServers(cser); err != nil {
		log.Panic(err)
	} else {
		log.Println("Inserted ", nser)
	}

}
