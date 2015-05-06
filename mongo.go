package main

import (
	"fmt"
	"log"
	"time"
	//  "strconv"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Item struct {
	Key string
	N   int
}

func main() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	id := "somekey.subkey"
	tstart := time.Now()
	c := session.DB("statsdb").C("pushstats")

	doc := Item{}
	for i := 0; i < 10000; i++ {
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"n": 1}},
			ReturnNew: true,
			Upsert:    true,
		}
		info, err := c.Find(bson.M{"key": id}).Apply(change, &doc)
		if err != nil {
			log.Fatal("apply", err)
		}
		if info != nil {
			fmt.Println(doc.N)
		}
	}
	ts := time.Now().Sub(tstart).Seconds()

	result := Item{}
	err = c.Find(bson.M{"key": &bson.RegEx{Pattern: ".*some.*", Options: "i"}}).One(&result)
	if err != nil {
		log.Fatal("one:", err)
	}

	//c.RemoveAll(nil);

	fmt.Println("result:", result)
	fmt.Println("time for exection:", ts)
}
