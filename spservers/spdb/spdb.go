package spdb

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	speedtestmongo = "mongodb://localhost:27017"
	Colserver      = "speedserver"
	Collinks       = "links"
	Coldatastatus  = "datastatus"
)

type SpeedtestMongo struct {
	config   *DBConfig
	Client   *mongo.Client
	Database *mongo.Database
}

type DBConfig struct {
	Username string
	Password string
	AuthDB   string
	DB       string
}

func connectmongo(mongopath, dbname string) (*mongo.Client, *mongo.Database) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongopath))
	if err != nil {
		log.Panic(err)
		return nil, nil
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Panic(err)
		return nil, nil
	}
	log.Println("Database connected")
	return client, client.Database(dbname)
}

func (cm *SpeedtestMongo) Close() {
	if cm.Client != nil {
		err := cm.Client.Disconnect(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Database connection ends")
	} else {
		log.Println("Mongo clientt is nil")
	}
}

func NewMongoDB(config string, dbname string) *SpeedtestMongo {
	cfile, err := os.Open(config)
	defer cfile.Close()
	if err != nil {
		log.Panic(err)
		return nil
	}
	c := &SpeedtestMongo{}
	decoder := json.NewDecoder(cfile)
	c.config = &DBConfig{}
	err = decoder.Decode(c.config)
	if err != nil {
		log.Panic(err)
		return nil
	}
	mongostr := "mongodb://"
	if c.config.DB == "" {
		log.Fatal("no database was selected")
	}
	if c.config.Username != "" && c.config.Password != "" {
		mongostr = mongostr + c.config.Username + ":" + c.config.Password + "@" + c.config.DB
		if c.config.AuthDB != "" {
			mongostr = mongostr + "/?authSource=" + c.config.AuthDB
		}
	}
	c.Client, c.Database = connectmongo(mongostr, dbname)
	return c
}

func (cm *SpeedtestMongo) ResetEnable(servertype string) {
	if cm.Database != nil {
		cspeed := cm.Database.Collection(Colserver)
		filter := bson.D{{"type", servertype}}
		update := bson.D{{"$set", bson.D{{"enabled", false}}}}
		res, err := cspeed.UpdateMany(context.TODO(), filter, update)
		if err != nil {
			log.Fatal(err)
		}
		if res.MatchedCount > 0 {
			log.Println("Distabled", res.MatchedCount, servertype, "servers")
		}
	}
}

func (cm *SpeedtestMongo) InsertServers(servers []SpeedServer) (int, error) {
	if cm.Database != nil && len(servers) > 0 {
		nodoc := 0
		cspeed := cm.Database.Collection(Colserver)
		opt := options.FindOneAndReplace().SetUpsert(true)
		for _, ser := range servers {
			filter := bson.D{{"type", ser.Type}, {"id", ser.Id}}
			res := cspeed.FindOneAndReplace(context.TODO(), filter, ser, opt)
			if res.Err() != nil {
				if res.Err() == mongo.ErrNoDocuments {
					nodoc++
					//	log.Println("Server not found", ser.Type, ser.Id)
				} else {
					log.Fatal(res.Err())
				}
			}
		}

		return nodoc, nil
	}
	if cm.Database == nil {
		return 0, errors.New("Database is nil")
	}
	return 0, nil
}

func (cm *SpeedtestMongo) QueryServersRaw(filters interface{}) (*mongo.Cursor, error) {
	if cm.Database != nil {
		cspeed := cm.Database.Collection(Colserver)
		return cspeed.Find(context.TODO(), filters)
	} else {
		return nil, errors.New("Database is nil")
	}
}

func (cm *SpeedtestMongo) QueryEnabledServersbyType(stype string) ([]SpeedServer, error) {
	var allservers []SpeedServer
	filter := bson.D{{"type", stype}, {"enabled", true}}
	cursor, err := cm.QueryServersRaw(filter)
	err = cursor.All(context.TODO(), &allservers)
	return allservers, err
}

func (cm *SpeedtestMongo) QueryEnabledServers() ([]SpeedServer, error) {
	var allservers []SpeedServer
	filter := bson.D{{"enabled", true}}
	cursor, err := cm.QueryServersRaw(filter)
	err = cursor.All(context.TODO(), &allservers)
	return allservers, err
}

func (cm *SpeedtestMongo) QueryServersbyIPv4(serverip net.IP) ([]SpeedServer, error) {
	var allservers []SpeedServer
	filter := bson.D{{"ipv4", serverip.String()}}
	cursor, err := cm.QueryServersRaw(filter)
	err = cursor.All(context.TODO(), &allservers)
	return allservers, err
}
func (cm *SpeedtestMongo) QueryDataStatus(mon string) (*VMDataStatus, error) {
	if cm.Database != nil {
		cdata := cm.Database.Collection(Coldatastatus)
		filter := bson.D{{"mon", mon}}
		var status VMDataStatus
		var err error
		res := cdata.FindOne(context.TODO(), filter)
		if res.Err() == mongo.ErrNoDocuments {
			return &VMDataStatus{}, nil
		} else {
			err = res.Decode(&status)
		}
		return &status, err
	} else {
		return nil, errors.New("Database is nil")
	}
}

func (cm *SpeedtestMongo) UpdateDataStatus(vmstatus *VMDataStatus) {
	cdata := cm.Database.Collection(Coldatastatus)
	opt := options.FindOneAndReplace().SetUpsert(true)
	filter := bson.D{{"mon", vmstatus.Mon}}
	res := cdata.FindOneAndReplace(context.TODO(), filter, vmstatus, opt)
	if res.Err() != nil {
		if res.Err() != mongo.ErrNoDocuments {
			//	log.Println("Server not found", ser.Type, ser.Id)
			log.Fatal(res.Err())
		}
	}

}

func (cm *SpeedtestMongo) UpdateLinkstoMongo(region string, linkmap map[string]*Link) {
	if cm != nil {
		//		basename := filepath.Base(LinkFile)
		//		namere := regexp.MustCompile(`(\S+)\.\d+\.links\.out`)
		//		namesplit := namere.FindStringSubmatch(LinkFile)
		//		region := namesplit[1]
		clink := cm.Database.Collection(Collinks)
		rmfilter := bson.D{{"region", region}}
		clink.DeleteMany(context.TODO(), rmfilter)
		for linkkey, link := range linkmap {
			link.Region = region
			link.Linkkey = linkkey
			_, err := clink.InsertOne(context.TODO(), link)
			if err != nil {
				log.Panic(err)
			}
		}
	}
}
