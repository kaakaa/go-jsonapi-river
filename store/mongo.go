package store

import (
	"log"

	"gopkg.in/mgo.v2"
)

type MongoCollection struct {
	Collection *mgo.Collection
}

func NewMongo(server, db, collections string) (*MongoCollection, error) {
	session, err := mgo.Dial(server)
	if err != nil {
		return nil, err
	}
	// defer session.Close()
	return &MongoCollection{Collection: session.DB(db).C(collections)}, nil
}

func (m *MongoCollection) Write(issue map[string]interface{}) {
	id := issue["_id"]
	q := m.Collection.FindId(id)
	n, err := q.Count()
	if err != nil {
		panic(err)
	}

	if n == 0 {
		log.Printf("Insert _id: %v", id)
		err = m.Collection.Insert(issue)
		if err != nil {
			panic(err)
		}
	} else if n == 1 {
		log.Printf("Duplicated _id: %v", id)
		err = m.Collection.UpdateId(id, issue)
		if err != nil {
			panic(err)
		}
	} else {
		log.Fatalf("Mongo has had already data with duplicate _id: %v", id)
	}
}
