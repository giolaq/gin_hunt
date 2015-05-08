package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/nichel/gin_hunt/models"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	HUNT_PATH     = "huntdata/"
	JSON_FILENAME = "sampleHunt.json"

	MONGO_DB         = "HUNT_DB"
	MONGO_COLLECTION = "HUNT_COLL"
)

type Response struct {
	Cached bool
	Items  interface{}
	Err    string
}

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
	debug := flag.Bool("d", false, "start in debug mode")
	port := flag.String("port", "8080", "port number")
	mongo_url := flag.String("mongod", "localhost", "mongodb url")

	flag.Parse()

	if *debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(MongoDB(*mongo_url))

	r.GET("/hunt", func(c *gin.Context) {
		var response Response

		var huntlist []*model.Hunt
		err := mDB.C(MONGO_COLLECTION).Find(nil).All(&huntlist)
		if err != nil {
			response.Err = err.Error()
		}

		for _, hunt := range huntlist {
			hunt.Clues = nil
		}

		response.Cached = false
		response.Items = huntlist

		c.JSON(http.StatusOK, response)
	})

	r.GET("/hunt/:hunt_id", func(c *gin.Context) {
		var response Response

		id := c.Params.ByName("hunt_id")

		var hunt model.Hunt
		err := mDB.C(MONGO_COLLECTION).Find(bson.M{"id": id}).One(&hunt)
		if err != nil {
			response.Err = err.Error()
		}

		response.Cached = false
		response.Items = hunt

		c.JSON(http.StatusOK, response)
	})

	admin := r.Group("/admin")
	admin.Use(gin.BasicAuth(gin.Accounts{"admin": "admin"}))

	admin.POST("/hunt", func(c *gin.Context) {
		var hunt model.Hunt

		if c.Bind(&hunt) {
			id := hunt.Id

			info, err := mDB.C(MONGO_COLLECTION).Upsert(bson.M{"id": id}, hunt)
			if err != nil {
				panic(err)
			}

			c.JSON(http.StatusCreated, info)
		}
	})

	admin.DELETE("/hunt", func(c *gin.Context) {
		info, err := mDB.C(MONGO_COLLECTION).RemoveAll(nil)
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, info)
	})

	admin.DELETE("/hunt/:hunt_id", func(c *gin.Context) {
		id := c.Params.ByName("hunt_id")

		info, err := mDB.C(MONGO_COLLECTION).RemoveAll(bson.M{"id": id})
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, info)
	})

	admin.HEAD("createZip/:hunt_id/:filename", func(c *gin.Context) {
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
	r.Static("web", "web/")

	r.Run(":" + (*port))
}
