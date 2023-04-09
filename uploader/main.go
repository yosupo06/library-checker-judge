package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/minio/minio-go/v6"
	"github.com/yosupo06/library-checker-judge/api/clientutil"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
)

func main() {
	dir := flag.String("dir", "../../library-checker-problems", "directory of library-checker-problems")

	apiHost := flag.String("apihost", "localhost:50051", "api host")
	apiUser := flag.String("apiuser", "upload", "api user")
	apiPass := flag.String("apipass", "password", "api password")

	pgHost := flag.String("pghost", "localhost", "postgre host")
	pgUser := flag.String("pguser", "postgres", "postgre user")
	pgPass := flag.String("pgpass", "passwd", "postgre password")
	pgTable := flag.String("pgtable", "librarychecker", "postgre table name")

	minioHost := flag.String("miniohost", "localhost:9000", "minio host")
	minioID := flag.String("minioid", "minio", "minio ID")
	minioKey := flag.String("miniokey", "miniopass", "minio access key")
	minioBucket := flag.String("miniobucket", "testcase", "minio bucket")

	useTLS := flag.Bool("tls", false, "use https for api / minio")

	t := flag.String("toml", "", "toml file of upload problem")

	flag.Parse()

	// connect db
	db := database.Connect(
		*pgHost,
		"5432",
		*pgTable,
		*pgUser,
		*pgPass,
		false)

	if *t == "" {
		log.Fatal("Please specify toml")
	}

	p := problem{
		root: *dir,
		base: path.Dir(*t),
		name: path.Base(path.Dir(*t)),
	}
	toml.DecodeFile(*t, &p.info)

	// connect minio
	mc, err := minio.New(
		*minioHost, *minioID, *minioKey, *useTLS,
	)
	if err != nil {
		log.Fatal("Cannot connect to Minio:", err)
	}

	conn := clientutil.ApiConnect(*apiHost, *useTLS)
	client := pb.NewLibraryCheckerInternalServiceClient(conn)
	ctx := login(client, *apiUser, *apiPass)

	if err := upload(p, mc, *minioBucket, db); err != nil {
		log.Fatal("Failed to upload problem: ", err)
	}

	if err := uploadCategories(*dir, client, ctx); err != nil {
		log.Fatal("Failed to update categories: ", err)
	}
}

func upload(p problem, mc *minio.Client, bucket string, db *gorm.DB) error {
	log.Print("Upload: ", p.name)

	v, err := p.version()
	if err != nil {
		log.Fatal("Failed to fetch version: ", err)
	}
	log.Print("New Version: ", v)

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

	statement, err := ioutil.ReadFile(path.Join(p.base, "task_body.html"))
	if err != nil {
		return err
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
		return err
	}

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

func login(client pb.LibraryCheckerInternalServiceClient, user, password string) context.Context {
	ctx := context.Background()
	resp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     user,
		Password: password,
	})

	if err != nil {
		log.Fatal("Cannot login to API Server:", err)
	}
	return clientutil.ContextWithToken(ctx, resp.Token)
}

func uploadCategories(dir string, client pb.LibraryCheckerInternalServiceClient, ctx context.Context) error {
	var data struct {
		Categories []struct {
			Name     string
			Problems []string
		}
	}
	if _, err := toml.DecodeFile(path.Join(dir, "categories.toml"), &data); err != nil {
		log.Fatal(err)
	}

	req := pb.ChangeProblemCategoriesRequest{}

	for _, c := range data.Categories {
		req.Categories = append(req.Categories,
			&pb.ProblemCategory{
				Title:    c.Name,
				Problems: c.Problems,
			})
	}

	_, err := client.ChangeProblemCategories(ctx, &req)
	return err
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
	sort.Strings(arr)
	copy(arr, hashes)

	h := sha256.New()
	for _, v := range arr {
		h.Write([]byte(v))
	}
	return fmt.Sprintf("%x", h.Sum([]byte{}))
}
