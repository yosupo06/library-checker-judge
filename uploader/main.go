package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
)

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

type FileInfo struct {
	base     string
	path     string
	required bool
}

func (p *problem) fileInfos() []FileInfo {
	return []FileInfo{
		// Common files
		// TODO: stop to manually add all common/*.h
		{
			base:     p.root,
			path:     path.Join("common", "fastio.h"),
			required: true,
		},
		{
			base:     p.root,
			path:     path.Join("common", "random.h"),
			required: true,
		},
		{
			base:     p.root,
			path:     path.Join("common", "testlib.h"),
			required: true,
		},
		// Problem files
		{
			base:     p.base,
			path:     path.Join("task.md"),
			required: true,
		},
		{
			base:     p.base,
			path:     path.Join("info.toml"),
			required: true,
		},
		{
			base:     p.base,
			path:     path.Join("checker.cpp"),
			required: true,
		},
		{
			base:     p.base,
			path:     path.Join("params.h"),
			required: true,
		},
		// for C++(Function)
		{
			base:     p.base,
			path:     path.Join("grader", "grader.cpp"),
			required: false,
		},
		{
			base:     p.base,
			path:     path.Join("grader", "solve.hpp"),
			required: false,
		},
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

func (p *problem) version() (string, error) {
	hashes := []string{}

	if h, err := p.testCasesHash(); err != nil {
		return "", err
	} else {
		hashes = append(hashes, h)
	}

	for _, info := range p.fileInfos() {
		path := path.Join(info.base, info.path)
		h, err := fileHash(path)
		if info.required && err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}

	return joinHashes(hashes), nil
}

func (p *problem) testCasesHash() (string, error) {
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

func (p *problem) uploadTestcases(mc *minio.Client, bucket string, publicBucket string) error {
	h, err := p.testCasesHash()
	if err != nil {
		return err
	}
	v, err := p.version()
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "testcase*.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	gzipWriter := gzip.NewWriter(tempFile)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, ext := range []string{"in", "out"} {
		if err := filepath.Walk(path.Join(p.base, ext), func(fpath string, info fs.FileInfo, err error) error {
			if strings.Contains(fpath, "example") {
				if _, err := mc.FPutObject(context.Background(), publicBucket, fmt.Sprintf("v2/%s/%s/%s/%s", p.name, v, ext, path.Base(fpath)), fpath, minio.PutObjectOptions{}); err != nil {
					return err
				}
			}

			if path.Ext(fpath) == fmt.Sprintf(".%s", ext) {
				file, err := os.Open(fpath)
				if err != nil {
					return err
				}
				defer file.Close()

				fileInfo, err := file.Stat()
				if err != nil {
					return err
				}

				header := &tar.Header{
					Name: fmt.Sprintf("%s/%s", ext, filepath.Base(fpath)),
					Size: fileInfo.Size(),
					Mode: 0600,
				}

				if err := tarWriter.WriteHeader(header); err != nil {
					return err
				}

				_, err = io.Copy(tarWriter, file)
				if err != nil {
					return err
				}

				return nil
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}

	if _, err := tempFile.Seek(0, 0); err != nil {
		return err
	}
	fileInfo, err := tempFile.Stat()
	if err != nil {
		return err
	}

	if _, err := mc.PutObject(context.Background(), bucket, fmt.Sprintf("v2/%s/%s.tar.gz", p.name, h), tempFile, fileInfo.Size(), minio.PutObjectOptions{}); err != nil {
		return err
	}

	return nil
}

func (p *problem) uploadFiles(mc *minio.Client, bucket string) error {
	v, err := p.version()
	if err != nil {
		log.Fatal("Failed to fetch version: ", err)
	}

	for _, info := range p.fileInfos() {
		src := path.Join(info.base, info.path)
		if _, err := os.Stat(src); err != nil {
			if info.required {
				return errors.New(fmt.Sprintf("required file: %s/%s not found", info.base, info.path))
			}
			continue
		}

		if _, err := mc.FPutObject(context.Background(), bucket, fmt.Sprintf("v2/%s/%s/%s", p.name, v, info.path), src, minio.PutObjectOptions{}); err != nil {
			return err
		}
	}

	return nil
}

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
		*minioHost, &minio.Options{
			Creds:  credentials.NewStaticV4(*minioID, *minioKey, ""),
			Secure: *useTLS,
		},
	)
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

		v, err := p.version()
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
		dbP.SourceUrl = fmt.Sprintf("https://github.com/yosupo06/library-checker-problems/tree/master/%v/%v", path.Base(path.Dir(p.base)), p.name)

		oldV := dbP.Version
		if newV, err := p.version(); err != nil {
			log.Fatalln("Failed to calculate problem version:", err)
		} else {
			dbP.Version = newV
		}

		oldH := dbP.TestCasesVersion
		if newH, err := p.testCasesHash(); err != nil {
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

			if err := p.uploadTestcases(mc, *minioBucket, *minioPublicBucket); err != nil {
				log.Fatalln("Failed to upload testcases:", err)
			}
		} else {
			log.Println("Skip test cases uploading")
		}

		if versionUpdated || *forceUpload {
			if err := p.uploadFiles(mc, *minioPublicBucket); err != nil {
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
