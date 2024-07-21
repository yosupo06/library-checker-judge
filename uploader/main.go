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

type problem struct {
	dir  storage.ProblemDir
	name string

	info struct {
		Title     string
		TimeLimit float64
	}
}

func newProblem(rootDir, tomlPath string) (*problem, error) {
	baseDir := path.Dir(tomlPath)
	p := problem{
		dir: storage.ProblemDir{
			Name: path.Base(baseDir),
			Root: rootDir,
			Base: baseDir,
		},
		name: path.Base(baseDir),
	}

	if _, err := toml.DecodeFile(tomlPath, &p.info); err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *problem) generate() error {
	cmd := exec.Command(path.Join(p.dir.Root, "generate.py"), "-p", p.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *problem) clean() error {
	cmd := exec.Command(path.Join(p.dir.Root, "generate.py"), "--clean", "-p", p.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	dir := flag.String("dir", "../../library-checker-problems", "directory of library-checker-problems")

	minioBucket := flag.String("miniobucket", "testcase", "minio bucket")
	minioPublicBucket := flag.String("miniopublicbucket", "testcase-public", "minio public bucket")

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

	db := database.Connect(database.GetDSNFromEnv(), false)

	// connect minio
	mc, err := storage.Connect(storage.GetConfigFromEnv())
	if err != nil {
		log.Fatalln("Cannot connect to Minio:", err)
	}

	for _, t := range tomls {
		p, err := newProblem(*dir, t)
		if err != nil {
			log.Fatalln("Failed to fetch problem info:", err)
		}

		log.Println("Upload problem:", p.name)

		// clean testcase & generate params.h
		if err := p.clean(); err != nil {
			log.Fatalln("Failed to clean:", err)
		}

		v, err := p.dir.Version()
		if err != nil {
			log.Fatalln("Failed to calculate version:", err)
		}
		log.Println("Problem version:", v)

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
		dbP.SourceUrl = fmt.Sprintf("https://github.com/yosupo06/library-checker-problems/tree/master/%v/%v", path.Base(path.Dir(p.dir.Base)), p.name)

		oldV := dbP.Version
		if newV, err := p.dir.Version(); err != nil {
			log.Fatalln("Failed to calculate problem version:", err)
		} else {
			dbP.Version = newV
		}

		oldH := dbP.TestCasesVersion
		if newH, err := p.dir.TestCaseHash(); err != nil {
			log.Fatalln("Failed to calculate test cases hash:", err)
		} else {
			dbP.TestCasesVersion = newH
		}

		versionUpdated := (dbP.Version != oldV)
		testcaseUpdated := (dbP.TestCasesVersion != oldH)

		if versionUpdated || *forceUpload {
			if err := p.generate(); err != nil {
				log.Fatalln("Failed to generate:", err)
			}

			if err := p.dir.UploadTestcases(mc, *minioBucket, *minioPublicBucket); err != nil {
				log.Fatalln("Failed to upload testcases:", err)
			}
		} else {
			log.Println("Skip test cases uploading")
		}

		if versionUpdated || *forceUpload {
			if err := p.dir.UploadFiles(mc, *minioPublicBucket); err != nil {
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

	if err := uploadCategories(*dir, db); err != nil {
		log.Fatal("Failed to update categories: ", err)
	}
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
