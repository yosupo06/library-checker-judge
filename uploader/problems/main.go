package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/storage"
)

func main() {
	problemsDir := flag.String("dir", "../../library-checker-problems", "directory of library-checker-problems")

	discordUrl := flag.String("discordwebhook", "", "webhook URL of discord")

	forceUpload := flag.Bool("force", false, "force upload even if the version is the same")

	flag.Parse()

	tomls := flag.Args()

	// connect discord
	var dc webhook.Client
	if *discordUrl != "" {
		c, err := webhook.NewWithURL(*discordUrl)
		if err != nil {
			slog.Error("Failed to init discord client", "err", err)
			os.Exit(1)
		}
		dc = c
	}

	// connect database
	db := database.Connect(database.GetDSNFromEnv(), false)

	// connect storage client
	storageClient, err := storage.Connect(storage.GetConfigFromEnv())
	if err != nil {
		slog.Error("Failed to connect to Minio", "err", err)
		os.Exit(1)
	}

	for _, t := range tomls {
		slog.Info("Upload problem", "toml", t)

		// clean testcase & generate params.h
		if err := clean(*problemsDir, t); err != nil {
			slog.Error("Failed to clean", "err", err)
			os.Exit(1)
		}

		// generate problem info
		target, err := storage.NewUploadTarget(path.Dir(t), *problemsDir)
		if err != nil {
			slog.Error("Failed to build UploadTarget", "err", err)
			os.Exit(1)
		}
		name := target.Problem.Name
		v := target.Problem.Version
		ov := target.Problem.OverallVersion
		h := target.Problem.TestCaseVersion
		slog.Info("Problem info", "name", name, "version", v, "overall_version", ov, "hash", h)

		// fetch problem info from database
		dbP, err := database.FetchProblem(db, name)
		newProblem := (err == database.ErrNotExist)
		if newProblem {
			slog.Info("New problem")
			dbP = database.Problem{
				Name: name,
			}
		} else if err != nil {
			slog.Error("Failed to fetch problem", "err", err)
			os.Exit(1)
		}

		// parse info.toml
		info, err := storage.ParseInfo(t)
		if err != nil {
			slog.Error("Failed to parse info.toml", "err", err)
			os.Exit(1)
		}

		versionUpdated := (v != dbP.Version)
		overallVersionUpdated := (ov != dbP.OverallVersion)
		testcaseUpdated := (h != dbP.TestCasesVersion)

		// update problem fields
		dbP.Title = info.Title
		dbP.Timelimit = int32(info.TimeLimit * 1000)
		dbP.SourceUrl = toSourceURL(t)
		dbP.Version = v
		dbP.OverallVersion = ov
		dbP.TestCasesVersion = h

		// upload test cases
		if testcaseUpdated || *forceUpload {
			if err := generate(*problemsDir, t); err != nil {
				slog.Error("Failed to generate", "err", err)
				os.Exit(1)
			}
			if err := target.UploadTestcases(storageClient); err != nil {
				slog.Error("Failed to upload test cases", "err", err)
				os.Exit(1)
			}
		} else {
			slog.Info("Skip to upload test cases")
		}

		// upload public files
		if versionUpdated || *forceUpload {
			if err := target.UploadPublicFilesV3(storageClient); err != nil {
				slog.Error("Failed to upload public files", "err", err)
				os.Exit(1)
			}
		}

		if overallVersionUpdated || *forceUpload {
			if err := target.UploadPublicFilesV4(storageClient); err != nil {
				slog.Error("Failed to upload public files (v4)", "err", err)
				os.Exit(1)
			}
		}
		if !(versionUpdated || overallVersionUpdated || *forceUpload) {
			slog.Info("Skip to upload public files")
		}

		if err := clean(*problemsDir, t); err != nil {
			slog.Error("Failed to clean", "err", err)
			os.Exit(1)
		}

		if err := database.SaveProblem(db, dbP); err != nil {
			slog.Error("Failed to upload problem info", "err", err)
			os.Exit(1)
		}

		if dc != nil && testcaseUpdated {
			if newProblem {
				if _, err := dc.CreateMessage(discord.NewWebhookMessageCreateBuilder().
					AddEmbeds(discord.NewEmbedBuilder().
						SetTitlef("New problem added: %s", info.Title).
						SetColor(0x00ff00).
						SetURLf("https://judge.yosupo.jp/problem/%s", name).
						AddField("Github", fmt.Sprintf("[link](%s)", dbP.SourceUrl), false).
						AddField("Test case hash", v[0:16], false).
						Build()).
					Build(),
				); err != nil {
					slog.Error("Failed to send message", "err", err)
				}
			} else {
				if _, err := dc.CreateMessage(discord.NewWebhookMessageCreateBuilder().
					AddEmbeds(discord.NewEmbedBuilder().
						SetTitlef("Testcase updated: %s", info.Title).
						SetColor(0x0000ff).
						SetURLf("https://judge.yosupo.jp/problem/%s", name).
						AddField("Github", fmt.Sprintf("[link](%s)", dbP.SourceUrl), false).
						AddField("New test case hash", v[0:16], false).
						Build()).
					Build(),
				); err != nil {
					slog.Error("Failed to send message", "err", err)
				}
			}
		}
	}

	// Note: Category upload is handled by separate CLI: ./categories
}

func generate(problemsDir, tomlPath string) error {
	cmd := exec.Command(path.Join(problemsDir, "generate.py"), tomlPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func clean(problemsDir, tomlPath string) error {
	cmd := exec.Command(path.Join(problemsDir, "generate.py"), "--clean", tomlPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func toSourceURL(tomlPath string) string {
	dir := path.Dir(tomlPath)
	return fmt.Sprintf("https://github.com/yosupo06/library-checker-problems/tree/master/%v/%v", path.Base(path.Dir(dir)), path.Base(dir))
}
