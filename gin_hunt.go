package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os/exec"
)

type Tag struct {
	Id string `bson:"id" json:"id"`
}

type Question struct {
	Question      string   `bson:"question" json:"question"`
	Answers       []string `bson:"anwsers" json:"anwsers"`
	CorrectAnswer int      `bson:"correctAnswer" json:"correctAnswer"`
	WrongMsg      string   `bson:"wrongMessage" json:"wrongMessage"`
	RightMsg      string   `bson:"rightMessage" json:"rightMessage"`
}

type Clue struct {
	Type         string   `bson:"type" json:"type"`
	Id           string   `bson:"id" json:"id"`
	ShuffleGroup int      `bson:"shufflegroup" json:"shufflegroup"`
	DisplayName  string   `bson:"displayName" json:"displayName"`
	DisplayText  string   `bson:"displayText" json:"displayText"`
	DisplayImage string   `bson:"displayImage" json:"displayImage"`
	Tags         []Tag    `bson:"tags" json:"tags"`
	Questions    Question `bson:"question" json:"question"`
}

type Hunt struct {
	Type        string `bson:"type" json:"type"`
	DisplayName string `bson:"displayName" json:"displayName"`
	Id          string `bson:"id" json:"id"`
	Clues       []Clue `bson:"clues" json:"clues"`
}

const (
	HUNT_PATH     = "huntdata/"
	JSON_FILENAME = "sampleHunt.json"
	ZIP_FILENAME  = "hunt.zip"

	MONGO_URL        = "localhost"
	MONGO_DB         = "mydb"
	MONGO_COLLECTION = "hunt"
)

func DB() gin.HandlerFunc {
	session, err := mgo.Dial(MONGO_URL)
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		s := session.Clone()
		s.DB(MONGO_DB)
		defer s.Close()

		c.Set(MONGO_DB, s.DB(MONGO_DB))
		c.Next()
	}
}

func main() {
	r := gin.Default()

	r.Use(DB())

	r.POST("/hunt", func(c *gin.Context) {
		var hunt Hunt
		ok := c.Bind(&hunt)

		if ok == true {
			db := c.MustGet(MONGO_DB).(*mgo.Database)
			db.C(MONGO_COLLECTION).Insert(hunt)
		}
	})

	r.GET("/hunt", func(c *gin.Context) {
		var hunt []Hunt

		db := c.MustGet(MONGO_DB).(*mgo.Database)
		db.C(MONGO_COLLECTION).Find(nil).All(&hunt)

		c.JSON(200, hunt)
	})

	r.GET("/hunt/:hunt_id", func(c *gin.Context) {
		id := c.Params.ByName("hunt_id")

		var hunt Hunt

		db := c.MustGet(MONGO_DB).(*mgo.Database)
		db.C(MONGO_COLLECTION).Find(bson.M{"id": id}).One(&hunt)

		c.JSON(200, hunt)
	})

	r.PUT("/hunt/:hunt_id", func(c *gin.Context) {
		id := c.Params.ByName("hunt_id")

		var hunt Hunt
		ok := c.Bind(&hunt)

		if ok == true {
			db := c.MustGet(MONGO_DB).(*mgo.Database)
			db.C(MONGO_COLLECTION).Update(bson.M{"id": id}, hunt)
		}
	})

	r.DELETE("/hunt", func(c *gin.Context) {
		db := c.MustGet(MONGO_DB).(*mgo.Database)
		db.C(MONGO_COLLECTION).RemoveAll(nil)
	})

	r.DELETE("/hunt/:hunt_id", func(c *gin.Context) {
		id := c.Params.ByName("hunt_id")

		db := c.MustGet(MONGO_DB).(*mgo.Database)
		db.C(MONGO_COLLECTION).Remove(bson.M{"id": id})
	})

	r.PUT("/clue/:hunt_id", func(c *gin.Context) {
		hid := c.Params.ByName("hunt_id")

		var hunt Hunt
		db := c.MustGet(MONGO_DB).(*mgo.Database)
		db.C(MONGO_COLLECTION).Find(bson.M{"id": hid}).One(&hunt)

		var clue Clue
		c.Bind(&hunt)

		hunt.Clues = append(hunt.Clues, clue)

		db.C(MONGO_COLLECTION).Update(nil, hunt)
	})

	r.HEAD("createZip/:hunt_id", func(c *gin.Context) {
		hid := c.Params.ByName("hunt_id")

		var hunt Hunt
		db := c.MustGet(MONGO_DB).(*mgo.Database)
		db.C(MONGO_COLLECTION).Find(bson.M{"id": hid}).One(&hunt)

		buffer, err := json.MarshalIndent(hunt, "", "\t")
		if err != nil {
			c.Fail(400, err)
			return
		}

		err = ioutil.WriteFile(HUNT_PATH+JSON_FILENAME, buffer, 0666) //TODO err?
		if err != nil {
			c.Fail(400, err)
			return
		}

		err = exec.Command("zip", "-r", ZIP_FILENAME).Run()
		if err != nil {
			c.Fail(402, err)
			return
		}
	})

	r.GET(ZIP_FILENAME, func(c *gin.Context) {
		c.File(ZIP_FILENAME)
	})

	r.Static("data", HUNT_PATH)

	r.Run(":8080")
}
