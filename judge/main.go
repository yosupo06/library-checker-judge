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
	defer downloader.Close()

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
		case database.JUDGE_SUBMISSION:
			if err := execSubmissionTask(db, downloader, taskID, taskData.Submission); err != nil {
				slog.Error("failed to judge Submission", "err", err)
				continue
			}
		case database.JUDGE_HACK:
			if err := execHackTask(db, downloader, taskID, taskData.Hack); err != nil {
				slog.Error("failed to judge Hack", "err", err)
				continue
			}
		}
		database.FinishTask(db, taskID)
	}
}
