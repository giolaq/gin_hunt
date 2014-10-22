package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os/exec"
	"strconv"
)

type Tag struct {
	Id string `json:"id"`
}

type Question struct {
	Question      string   `json:"question"`
	Answers       []string `json:"anwsers"`
	CorrectAnswer int      `json:"correctAnswer"`
	WrongMsg      string   `json:"wrongMessage"`
	RightMsg      string   `json:"rightMessage"`
}

type Clue struct {
	Type         string   `json:"type"`
	Id           string   `json:"id"`
	ShuffleGroup int      `json:"shufflegroup"`
	DisplayName  string   `json:"displayName"`
	DisplayText  string   `json:"displayText"`
	DisplayImage string   `json:"displayImage"`
	Tags         []Tag    `json:"tags"`
	Questions    Question `json:"question"`
}

type Hunt struct {
	Type        string `json:"type"`
	DisplayName string `json:"displayName"`
	Id          string `json:"id"`
	Clues       []Clue `json:"clues"`
}

var (
	hunt Hunt
)

const (
	DATA_PATH = "huntdata"
)

func main() {
	print()

	r := gin.Default()

	r.GET("/hunt", func(c *gin.Context) {
		c.JSON(200, hunt)
	})

	r.POST("/hunt", func(c *gin.Context) {
		ok := c.Bind(&hunt)

		if ok == false {
			c.Abort(400)
		}
	})

	r.GET("/clue/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Params.ByName("id"))

		if err == nil && id < len(hunt.Clues) {
			clue := &(hunt.Clues[id])

			c.JSON(200, clue)
		}
	})

	r.POST("/clue", func(c *gin.Context) {
		var clue Clue
		ok := c.Bind(&clue)

		if ok == false {
			c.Abort(400)
		} else {
			hunt.Clues = append(hunt.Clues, clue)
		}
	})

	r.GET("hunt.zip", func(c *gin.Context) {
		buffer, _ := json.MarshalIndent(hunt, "", "\t")
		ioutil.WriteFile(DATA_PATH+"/hunt.json", buffer, 0666)

		exec.Command("zip", "-r", "hunt.zip", DATA_PATH).Run()

		c.File("hunt.zip")
	})

	r.Run(":8080")
}
