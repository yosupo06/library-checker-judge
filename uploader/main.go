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
	"strings"

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
	minioPublicBucket := flag.String("miniopublicbucket", "testcase-public", "minio public bucket")

	discordUrl := flag.String("discordwebhook", "", "webhook URL of discord")

	useTLS := flag.Bool("tls", false, "use https for api / minio")

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
		p, err := newProblem(*dir, t)
		if err != nil {
			log.Fatalln("failed to fetch problem info:", err)
		}

		log.Println("upload problem:", p.name)

		// clean testcase & generate params.h
		if err := p.clean(); err != nil {
			log.Fatalln("failed to clean:", err)
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

		if dbP == nil {
			dbP = &database.Problem{}
		}

		// update problem fields
		statement, err := ioutil.ReadFile(path.Join(p.base, "task_body.html"))
		if err != nil {
			log.Fatalln("failed to read task_body:", err)
		}
		dbP.Name = p.name
		dbP.Title = p.info.Title
		dbP.Timelimit = int32(p.info.TimeLimit * 1000)
		dbP.Statement = string(statement)
		dbP.SourceUrl = fmt.Sprintf("https://github.com/yosupo06/library-checker-problems/tree/master/%v/%v", path.Base(path.Dir(p.base)), p.name)
		dbP.Testhash = v

		if pV, err := p.publicFilesVersion(); err != nil {
			log.Fatalln("failed to calc public files hash:", err)
		} else {
			dbP.PublicFilesHash = pV
		}

		uploadData := (v != oldV || *forceUpload)
		if uploadData {
			if err := p.generate(); err != nil {
				log.Fatalln("failed to generate:", err)
			}

			if err := p.uploadFiles(mc, *minioBucket); err != nil {
				log.Fatalln("failed to upload files:", err)
			}
		} else {
			log.Println("version is the same, skip upload")
		}

		if err := p.uploadPublicFiles(mc, *minioPublicBucket); err != nil {
			log.Fatalln("failed to upload public files:", err)
		}

		if err := p.deleteFiles(mc, *minioBucket); err != nil {
			log.Fatalln("failed to clean minio:", err)
		}

		if err := p.clean(); err != nil {
			log.Fatalln("failed to clean local:", err)
		}

		if err := database.SaveProblem(db, *dbP); err != nil {
			log.Fatalln("failed to update problem info:", err)
		}

		if uploadData {
			if dc != nil {
				if oldV == "" {
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
						log.Fatal("error sending message:", err)
					}
				} else if oldV != v {
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
						log.Fatal("error sending message:", err)
					}
				}
			}
		}
	}

	if err := uploadCategories(*dir, db); err != nil {
		log.Fatal("Failed to update categories: ", err)
	}
}

type problem struct {
	root string
	base string
	name string
	v    string

	info struct {
		Title     string
		TimeLimit float64
	}
}

func newProblem(rootDir, tomlPath string) (*problem, error) {
	baseDir := path.Dir(tomlPath)
	p := problem{
		root: rootDir,
		base: baseDir,
		name: path.Base(baseDir),
	}

	if _, err := toml.DecodeFile(tomlPath, &p.info); err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *problem) generate() error {
	cmd := exec.Command(path.Join(p.root, "generate.py"), "--only-html", "-p", p.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *problem) clean() error {
	cmd := exec.Command(path.Join(p.root, "generate.py"), "--clean", "-p", p.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *problem) uploadFiles(mc *minio.Client, bucket string) error {
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

	return nil
}

func (p *problem) uploadPublicFiles(mc *minio.Client, bucket string) error {
	v, err := p.publicFilesVersion()
	if err != nil {
		log.Fatal("Failed to fetch version: ", err)
	}

	type Info struct {
		from string
		to   string
	}

	for _, info := range []Info{
		{from: path.Join(p.root, "common", "fastio.h"), to: fmt.Sprintf("v1/%v/%v/common/fastio.h", p.name, v)},
		{from: path.Join(p.base, "grader", "grader.cpp"), to: fmt.Sprintf("v1/%v/%v/grader/grader.cpp", p.name, v)},
		{from: path.Join(p.base, "grader", "solve.hpp"), to: fmt.Sprintf("v1/%v/%v/grader/solve.hpp", p.name, v)},
	} {
		if _, err := os.Stat(info.from); err != nil {
			continue
		}

		if _, err := mc.FPutObject(bucket, info.to, info.from, minio.PutObjectOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (p *problem) deleteFiles(mc *minio.Client, bucket string) error {
	v, err := p.version()
	if err != nil {
		log.Fatalln("failed to fetch version: ", err)
	}

	doneCh := make(chan struct{})
	defer close(doneCh)

	for object := range mc.ListObjects(bucket, fmt.Sprintf("v1/%v/", p.name), true, doneCh) {
		if strings.HasPrefix(object.Key, fmt.Sprintf("v1/%v/%v", p.name, v)) {
			continue
		}

		if err := mc.RemoveObject(bucket, object.Key); err != nil {
			log.Fatalln("failed to remove:", object.Key)
		}
	}
	return nil
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

func (p *problem) publicFilesVersion() (string, error) {
	hashes := []string{}

	for _, file := range []string{
		path.Join(p.root, "common", "fastio.h"),
		path.Join(p.base, "grader", "grader.cpp"),
		path.Join(p.root, "grader", "solve.h"),
	} {
		if _, err := os.Stat(file); err != nil {
			continue
		}

		if h, err := fileHash(file); err != nil {
			return "", err
		} else {
			hashes = append(hashes, h)
		}
	}

	return joinHashes(hashes), nil
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
