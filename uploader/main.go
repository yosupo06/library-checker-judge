package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/storage"
	"gorm.io/gorm"
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
			log.Fatal("Failed to init discord client:", err)
		}
		dc = c
	}

	// connect database
	db := database.Connect(database.GetDSNFromEnv(), false)

	// connect storage client
	storageClient, err := storage.Connect(storage.GetConfigFromEnv())
	if err != nil {
		log.Fatalln("Cannot connect to Minio:", err)
	}

	for _, t := range tomls {
		// clean testcase & generate params.h
		if err := clean(*problemsDir, t); err != nil {
			log.Fatalln("Failed to clean:", err)
		}

		p, err := newProblem(*problemsDir, t)
		if err != nil {
			log.Fatalln("Failed to fetch problem info:", err)
		}
		log.Println("Upload problem:", p.name)

		v := p.target.Problem.Version
		h := p.target.Problem.TestCaseHash
		log.Println("Problem version:", v, h)

		// clean testcase & generate params.h
		if err := p.clean(); err != nil {
			log.Fatalln("Failed to clean:", err)
		}

		dbP, err := database.FetchProblem(db, p.name)
		if err != nil {
			log.Fatalln("Failed to fetch problem:", err)
		}

		// update problem fields
		if dbP == nil {
			dbP = &database.Problem{}
		}
		dbP.Name = p.name
		dbP.Title = p.info.Title
		dbP.Timelimit = int32(p.info.TimeLimit * 1000)
		dbP.SourceUrl = toSourceURL(t)

		oldV := dbP.Version
		dbP.Version = v

		oldH := dbP.TestCasesVersion
		dbP.TestCasesVersion = h

		versionUpdated := (dbP.Version != oldV)
		testcaseUpdated := (dbP.TestCasesVersion != oldH)

		if versionUpdated || *forceUpload {
			if err := p.generate(); err != nil {
				log.Fatalln("Failed to generate:", err)
			}

			if err := p.target.UploadTestcases(storageClient); err != nil {
				log.Fatalln("Failed to upload testcases:", err)
			}
		} else {
			log.Println("Skip test cases uploading")
		}

		if versionUpdated || *forceUpload {
			if err := p.target.UploadFiles(storageClient); err != nil {
				log.Fatalln("Failed to upload public files:", err)
			}
		} else {
			log.Println("Skip public files uploading")
		}

		if err := p.clean(); err != nil {
			log.Fatalln("Failed to clean problem:", err)
		}

		if err := database.SaveProblem(db, *dbP); err != nil {
			log.Fatalln("Failed to update problem info:", err)
		}

		if dc != nil && testcaseUpdated {
			if oldH == "" {
				if _, err := dc.CreateMessage(discord.NewWebhookMessageCreateBuilder().
					AddEmbeds(discord.NewEmbedBuilder().
						SetTitlef("New problem added: %s", p.info.Title).
						SetColor(0x00ff00).
						SetURLf("https://judge.yosupo.jp/problem/%s", p.name).
						AddField("Github", fmt.Sprintf("[link](%s)", dbP.SourceUrl), false).
						AddField("Test case hash", v[0:16], false).
						Build()).
					Build(),
				); err != nil {
					log.Fatalln("Error sending message:", err)
				}
			} else if testcaseUpdated {
				if _, err := dc.CreateMessage(discord.NewWebhookMessageCreateBuilder().
					AddEmbeds(discord.NewEmbedBuilder().
						SetTitlef("Testcase updated: %s", p.info.Title).
						SetColor(0x0000ff).
						SetURLf("https://judge.yosupo.jp/problem/%s", p.name).
						AddField("Github", fmt.Sprintf("[link](%s)", dbP.SourceUrl), false).
						AddField("Old test case hash", oldV[0:16], false).
						AddField("New test case hash", v[0:16], false).
						Build()).
					Build(),
				); err != nil {
					log.Fatalln("Error sending message:", err)
				}
			}
		}
	}

	if err := uploadCategories(*problemsDir, db); err != nil {
		log.Fatal("Failed to update categories: ", err)
	}
}

type problem struct {
	target storage.UploadTarget
	name   string

	info struct {
		Title     string
		TimeLimit float64
	}
}

func newProblem(rootDir, tomlPath string) (*problem, error) {
	baseDir := path.Dir(tomlPath)

	target, err := storage.NewUploadTarget(baseDir, rootDir)
	if err != nil {
		return nil, nil
	}

	p := problem{
		target: target,
		name:   path.Base(baseDir),
	}

	if _, err := toml.DecodeFile(tomlPath, &p.info); err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *problem) generate() error {
	cmd := exec.Command(path.Join(p.target.Root, "generate.py"), "-p", p.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *problem) clean() error {
	cmd := exec.Command(path.Join(p.target.Root, "generate.py"), "--clean", "-p", p.name)
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

func uploadCategories(dir string, db *gorm.DB) error {
	var data struct {
		Categories []struct {
			Name     string
			Problems []string
		}
	}
	if _, err := toml.DecodeFile(path.Join(dir, "categories.toml"), &data); err != nil {
		log.Fatal(err)
	}

	cs := []database.ProblemCategory{}

	for _, c := range data.Categories {
		cs = append(cs,
			database.ProblemCategory{
				Title:    c.Name,
				Problems: c.Problems,
			})
	}

	return database.SaveProblemCategories(db, cs)
}

func toSourceURL(tomlPath string) string {
	dir := path.Dir(tomlPath)
	return fmt.Sprintf("https://github.com/yosupo06/library-checker-problems/tree/master/%v/%v", path.Base(path.Dir(dir)), path.Base(dir))
}
