package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/storage"
)

const POOLING_PERIOD = 3 * time.Second

func main() {
	flag.Parse()

	// connect db
	db := database.Connect(database.GetDSNFromEnv(), false)

	storageClient, err := storage.Connect(storage.GetConfigFromEnv())
	if err != nil {
		slog.Error("Failed to connect to storage", "err", err)
		os.Exit(1)
	}
	downloader, err := storage.NewTestCaseDownloader(storageClient)
	if err != nil {
		slog.Error("Failed to create TestCaseDownloader", "err", err)
		os.Exit(1)
	}
	defer func() { _ = downloader.Close() }()

	slog.Info("Start pooling")
	for {
		taskID, taskData, err := database.PopTask(db)
		if err != nil {
			slog.Error("PopJudgeTask failed", "err", err)
			time.Sleep(POOLING_PERIOD)
			continue
		}
		if taskID == -1 {
			time.Sleep(POOLING_PERIOD)
			continue
		}

		slog.Info("Start task", "ID", taskID)
		switch taskData.TaskType {
		case database.JudgeSubmission:
			submissionData, ok := taskData.Data.(database.SubmissionData)
			if !ok {
				slog.Error("Failed to cast to SubmissionData", "taskID", taskID)
				continue
			}
			if err := execSubmissionTask(db, downloader, taskID, submissionData); err != nil {
				slog.Error("Failed to judge Submission", "taskID", taskID, "err", err)
				continue
			}
		case database.JudgeHack:
			hackData, ok := taskData.Data.(database.HackData)
			if !ok {
				slog.Error("Failed to cast to HackData", "taskID", taskID)
				continue
			}
			if err := execHackTask(db, downloader, taskID, hackData.ID); err != nil {
				slog.Error("Failed to judge Hack", "taskID", taskID, "err", err)
				continue
			}
		}
		slog.Info("Finish task", "ID", taskID)
		_ = database.FinishTask(db, taskID)
	}
}
