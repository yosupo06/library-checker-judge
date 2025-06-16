package main

import (
	"log"

	"github.com/yosupo06/library-checker-judge/database"
)

var (
	rejudgeCmd         = app.Command("r", "Rejudge")
	rejudgeSubmissions = rejudgeCmd.Arg("id", "Submission ID").Required().Int32List()
)

func execRejudgeCmd() {
	db := database.Connect(database.GetDSNFromEnv(), true)

	for _, id := range *rejudgeSubmissions {
		log.Print("rejudge:", id)
		if err := database.PushSubmissionTask(db, database.SubmissionData{
			ID: id,
		}, 45); err != nil {
			log.Print("rejudge failed:", err)
		}
	}
}
