package main

import (
	"encoding/json"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/nichel/gin_hunt/models"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os/exec"
)

const (
	HUNT_PATH     = "huntdata/"
	JSON_FILENAME = "sampleHunt.json"

	DEF_MONGO_URL    = "localhost"
	MONGO_DB         = "HUNT_DB"
	MONGO_COLLECTION = "HUNT_COLL"
)

var (
	mDB *mgo.Database
)

func MongoDB(mongo_url string) gin.HandlerFunc {
	session, err := mgo.Dial(mongo_url)
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		s := session.Clone()
		mDB = s.DB(MONGO_DB)
		defer s.Close()

		c.Next()
	}
}

func main() {
	debug := flag.Bool("debug", false, "start in debug mode")
	port := flag.String("port", "8080", "port number")
	mongo_url := flag.String("mongod", DEF_MONGO_URL, "mongodb url")

	flag.Parse()

	if *debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(MongoDB(*mongo_url))

	r.POST("/hunt", func(c *gin.Context) {
		var hunt model.Hunt

		if c.Bind(&hunt) {
			id := hunt.Id

			info, err := mDB.C(MONGO_COLLECTION).Upsert(bson.M{"id": id}, hunt)
			if err != nil {
				panic(err)
			}

			c.JSON(200, info)
		}
	})

	r.GET("/hunt", func(c *gin.Context) {
		var hunt []model.Hunt

		err := mDB.C(MONGO_COLLECTION).Find(nil).All(&hunt)
		if err != nil {
			panic(err)
		}

		c.JSON(200, hunt)
	})

	r.GET("/hunt/:hunt_id", func(c *gin.Context) {
		id := c.Params.ByName("hunt_id")

		var hunt model.Hunt

		err := mDB.C(MONGO_COLLECTION).Find(bson.M{"id": id}).One(&hunt)
		if err != nil {
			panic(err)
		}

		c.JSON(200, hunt)
	})

	r.DELETE("/hunt", func(c *gin.Context) {
		info, err := mDB.C(MONGO_COLLECTION).RemoveAll(nil)
		if err != nil {
			panic(err)
		}

		c.JSON(200, info)
	})

	r.DELETE("/hunt/:hunt_id", func(c *gin.Context) {
		id := c.Params.ByName("hunt_id")

		info, err := mDB.C(MONGO_COLLECTION).RemoveAll(bson.M{"id": id})
		if err != nil {
			panic(err)
		}

		c.JSON(200, info)
	})

	r.PUT("/clue/:hunt_id", func(c *gin.Context) {
		id := c.Params.ByName("hunt_id")

		var hunt model.Hunt

		err := mDB.C(MONGO_COLLECTION).Find(bson.M{"id": id}).One(&hunt)
		if err != nil {
			panic(err)
		}

		var clue model.Clue

		if c.Bind(&clue) {
			hunt.Clues = append(hunt.Clues, clue)

			info, err := mDB.C(MONGO_COLLECTION).UpdateAll(bson.M{"id": id}, hunt)
			if err != nil {
				panic(err)
			}

			c.JSON(200, info)
		}
	})

	r.HEAD("createZip/:hunt_id/:filename", func(c *gin.Context) {
		id := c.Params.ByName("hunt_id")
		filename := c.Params.ByName("filename")

		var hunt model.Hunt

		err := mDB.C(MONGO_COLLECTION).Find(bson.M{"id": id}).One(&hunt)
		if err != nil {
			panic(err)
		}

		buffer, err := json.MarshalIndent(hunt, "", "\t")
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(HUNT_PATH+JSON_FILENAME, buffer, 0666)
		if err != nil {
			panic(err)
		}

		err = exec.Command("zip", "-r", "zip/"+filename, HUNT_PATH).Run()
		if err != nil {
			panic(err)
		}
	})

	r.Static("zip", "zip/")

	r.Run(":" + (*port))
}
