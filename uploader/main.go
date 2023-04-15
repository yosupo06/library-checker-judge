package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/minio/minio-go/v6"
	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
)

func main() {
	dir := flag.String("dir", "../../library-checker-problems", "directory of library-checker-problems")

	pgHost := flag.String("pghost", "localhost", "postgre host")
	pgUser := flag.String("pguser", "postgres", "postgre user")
	pgPass := flag.String("pgpass", "passwd", "postgre password")
	pgTable := flag.String("pgtable", "librarychecker", "postgre table name")

	minioHost := flag.String("miniohost", "localhost:9000", "minio host")
	minioID := flag.String("minioid", "minio", "minio ID")
	minioKey := flag.String("miniokey", "miniopass", "minio access key")
	minioBucket := flag.String("miniobucket", "testcase", "minio bucket")

	discordUrl := flag.String("discordwebhook", "", "webhook URL of discord")

	useTLS := flag.Bool("tls", false, "use https for api / minio")

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

	// connect db
	db := database.Connect(
		*pgHost,
		"5432",
		*pgTable,
		*pgUser,
		*pgPass,
		false)

	// connect minio
	mc, err := minio.New(
		*minioHost, *minioID, *minioKey, *useTLS,
	)
	if err != nil {
		log.Fatal("Cannot connect to Minio:", err)
	}

	for _, t := range tomls {
		p := problem{
			root: *dir,
			base: path.Dir(t),
			name: path.Base(path.Dir(t)),
		}
		log.Println("upload problem:", p.name)

		if _, err := toml.DecodeFile(t, &p.info); err != nil {
			log.Fatalln("failed to decode toml:", err)
		}

		v, err := p.version()
		if err != nil {
			log.Fatalln("failed to calculate version:", err)
		}
		log.Println("new version:", v)

		dbP, err := database.FetchProblem(db, p.name)
		if err != nil {
			log.Fatalln("failed to fetch problem:", err)
		}
		oldV := ""
		if dbP != nil {
			oldV = dbP.Testhash
		}
		if oldV == "" {
			log.Println("new problem")
		} else {
			log.Println("old version:", dbP.Testhash)
		}

		if v == oldV {
			log.Println("version is the same, skip")
			continue
		}

		if err := p.generate(); err != nil {
			log.Fatalln("failed to generate:", err)
		}

		if err := uploadFiles(p, mc, *minioBucket); err != nil {
			log.Fatalln("failed to upload files:", err)
		}

		statement, err := ioutil.ReadFile(path.Join(p.base, "task_body.html"))
		if err != nil {
			log.Fatalln("failed to read task_body:", err)
		}

		source := fmt.Sprintf("https://github.com/yosupo06/library-checker-problems/tree/master/%v/%v", path.Base(path.Dir(p.base)), p.name)

		if err := database.SaveProblem(db, database.Problem{
			Name:      p.name,
			Title:     p.info.Title,
			Timelimit: int32(p.info.TimeLimit * 1000),
			Statement: string(statement),
			SourceUrl: source,
			Testhash:  v,
		}); err != nil {
			log.Fatalln("failed to update problem info:", err)
		}

		if dc != nil {
			if oldV == "" {
				if _, err := dc.CreateMessage(discord.NewWebhookMessageCreateBuilder().
					AddEmbeds(discord.NewEmbedBuilder().
						SetTitlef("New Problem added: %s", p.info.Title).
						SetColor(0x00ff00).
						SetURLf("https://judge.yosupo.jp/problem/%s", p.name).
						AddField("Github", fmt.Sprintf("[link](%s)", source), false).
						AddField("Testcase hash", v, false).
						Build()).
					Build(),
				); err != nil {
					log.Fatal("error sending message:", err)
				}
			}
		}

		if err := p.clean(); err != nil {
			log.Fatalln("failed to clean:", err)
		}
	}

	if err := uploadCategories(*dir, db); err != nil {
		log.Fatal("Failed to update categories: ", err)
	}
}

func uploadFiles(p problem, mc *minio.Client, bucket string) error {
	v, err := p.version()
	if err != nil {
		log.Fatal("Failed to fetch version: ", err)
	}

	filepath.Walk(path.Join(p.base, "in"), func(fpath string, info fs.FileInfo, err error) error {
		if path.Ext(fpath) == ".in" {
			if _, err := mc.FPutObject(bucket, fmt.Sprintf("v1/%v/%v/testcases/in/%v", p.name, v, path.Base(fpath)), fpath, minio.PutObjectOptions{}); err != nil {
				return err
			}
		}
		return nil
	})
	filepath.Walk(path.Join(p.base, "out"), func(fpath string, info fs.FileInfo, err error) error {
		if path.Ext(fpath) == ".out" {
			if _, err := mc.FPutObject(bucket, fmt.Sprintf("v1/%v/%v/testcases/out/%v", p.name, v, path.Base(fpath)), fpath, minio.PutObjectOptions{}); err != nil {
				return err
			}
		}
		return nil
	})

	if _, err := mc.FPutObject(bucket, fmt.Sprintf("v1/%v/%v/checker.cpp", p.name, v), path.Join(p.base, "checker.cpp"), minio.PutObjectOptions{}); err != nil {
		return err
	}

	if _, err := mc.FPutObject(bucket, fmt.Sprintf("v1/%v/%v/include/params.h", p.name, v), path.Join(p.base, "params.h"), minio.PutObjectOptions{}); err != nil {
		return err
	}

	if _, err := mc.FPutObject(bucket, fmt.Sprintf("v1/%v/%v/include/testlib.h", p.name, v), path.Join(p.root, "common", "testlib.h"), minio.PutObjectOptions{}); err != nil {
		return err
	}

	if _, err := mc.FPutObject(bucket, fmt.Sprintf("v1/%v/%v/include/random.h", p.name, v), path.Join(p.root, "common", "random.h"), minio.PutObjectOptions{}); err != nil {
		return err
	}
	log.Print("File uploaded")

	return nil
}

type problem struct {
	root string
	base string
	name string

	info struct {
		Title     string
		TimeLimit float64
	}
}

func (p *problem) generate() error {
	cmd := exec.Command("../../library-checker-problems/generate.py", "--only-html", "-p", p.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *problem) clean() error {
	cmd := exec.Command("../../library-checker-problems/generate.py", "--clean", "-p", p.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *problem) checkerHash() (string, error) {
	return fileHash(path.Join(p.base, "checker.cpp"))
}

func (p *problem) caseHash() (string, error) {
	caseHash, err := ioutil.ReadFile(path.Join(p.base, "hash.json"))
	if err != nil {
		return "", err
	}
	var cases map[string]string
	if err := json.Unmarshal(caseHash, &cases); err != nil {
		return "", err
	}

	hashes := make([]string, 0, len(cases))
	for _, v := range cases {
		hashes = append(hashes, v)
	}
	return joinHashes(hashes), nil
}

func (p *problem) includeHash() (string, error) {
	hashes := []string{}

	for _, path := range []string{
		path.Join(p.root, "common", "testlib.h"),
		path.Join(p.root, "common", "random.h"),
		path.Join(p.base, "params.h"),
	} {
		h, err := fileHash(path)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}

	return joinHashes(hashes), nil
}

func (p *problem) version() (string, error) {
	hashes := []string{}

	if h, err := p.checkerHash(); err != nil {
		return "", err
	} else {
		hashes = append(hashes, h)
	}

	if h, err := p.caseHash(); err != nil {
		return "", err
	} else {
		hashes = append(hashes, h)
	}

	if h, err := p.includeHash(); err != nil {
		return "", err
	} else {
		hashes = append(hashes, h)
	}

	return joinHashes(hashes), nil
}

func fileHash(path string) (string, error) {
	checker, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(checker)), nil
}

func joinHashes(hashes []string) string {
	arr := make([]string, len(hashes))
	copy(arr, hashes)
	sort.Strings(arr)

	h := sha256.New()
	for _, v := range arr {
		h.Write([]byte(v))
	}
	return fmt.Sprintf("%x", h.Sum([]byte{}))
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
