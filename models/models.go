package model

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
	Id string `bson:"id" json:"id"`

	Type         string   `bson:"type" json:"type"`
	ShuffleGroup int      `bson:"shufflegroup" json:"shufflegroup"`
	DisplayName  string   `bson:"displayName" json:"displayName"`
	DisplayText  string   `bson:"displayText" json:"displayText"`
	DisplayImage string   `bson:"displayImage" json:"displayImage"`
	Tags         []Tag    `bson:"tags" json:"tags"`
	Questions    Question `bson:"question" json:"question"`
}

type Hunt struct {
	Id string `bson:"id" json:"id"`

	Type        string `bson:"type" json:"type"`
	DisplayName string `bson:"displayName" json:"displayName"`
	Clues       []Clue `bson:"clues" json:"clues"`
}
