package main

import (
	"log"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/yosupo06/library-checker-judge/database"
)

var (
	app                  = kingpin.New("rejudge", "Queue submissions for rejudge")
	rejudgeSubmissionIDs = app.Arg("id", "Submission ID").Required().Int32List()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	db := database.Connect(database.GetDSNFromEnv(), true)

	for _, id := range *rejudgeSubmissionIDs {
		log.Print("rejudge:", id)
		if err := database.PushSubmissionTask(db, database.SubmissionData{
			ID: id,
		}, 45); err != nil {
			log.Print("rejudge failed:", err)
		}
	}
}
