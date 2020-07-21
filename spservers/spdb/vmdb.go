package spdb

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

// VMCollection name of the vm collection
var VMCollection string = "vmInfo"

// VMInfo record containing information of VM
type VMInfo struct {
	Name string
	// Id instance id
	ID string
}

/* Mongo utils */

// InitMongoDB initialize a mongodb client
func InitMongoDB(mongoConfigFile string) *SpeedtestMongo {
	db := NewMongoDB(mongoConfigFile, "speedtest")
	if db.Client == nil {
		log.Fatal("Connect to mongodb err")
	}
	return db
}

// InsertInstanceIDName insert id and name of vm as a record to db
func InsertInstanceIDName(cm *SpeedtestMongo, VMs []VMInfo) {
	if cm.Database != nil && len(VMs) > 0 {
		collection := cm.Database.Collection(VMCollection)

		_, err := collection.Indexes().CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys:    bsonx.Doc{{"id", bsonx.Int32(1)}},
				Options: options.Index().SetUnique(true),
			},
		)
		handleError(err)

		// insert all VM in {name, instanceID} format
		interfaceSlice := make([]interface{}, len(VMs))
		for i, d := range VMs {
			interfaceSlice[i] = d
		}
		interfaceSlice = append(interfaceSlice, VMInfo{"test", "test"})

		opt := options.Update().SetUpsert(true)
		for _, ele := range VMs {
			//ricky: this query is not correct
			update := bson.D{{"$set", bson.D{{"id", ele.ID}, {"name", ele.Name}}}}
			filter := bson.D{{"id", ele.ID}}
			res, err := collection.UpdateOne(context.Background(), filter, update, opt)
			handleError(err)
			log.Println("updated:", res.ModifiedCount, "documents\n", "inserted ids:", res.UpsertedID)
		}
	}
}

// QueryVMIDByName query VM by its name, returning its id
func QueryVMIDByName(cm *SpeedtestMongo, name string) string {
	// create a value into which the result can be decoded
	var id string

	collection := cm.Database.Collection(VMCollection)
	filter := bson.D{{"name", name}}
	err := collection.FindOne(context.TODO(), filter).Decode(&id)
	handleError(err)

	log.Printf("Found a single document: %+v\n", id)
	return id
}

// ClearCollection delete everything inside a collection
func ClearCollection(collection mongo.Collection) {
	filter := bson.D{{}}
	res, err := collection.DeleteMany(context.Background(), filter)
	handleError(err)
	log.Println("Delete Result: ", res.DeletedCount)
}
