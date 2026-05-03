package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClientUploadDownloadWithEmulator(t *testing.T) {
	if os.Getenv("STORAGE_EMULATOR_HOST") == "" {
		t.Skip("STORAGE_EMULATOR_HOST is not set")
	}
	t.Setenv("STORAGE_PROJECT_ID", "dev-library-checker-project")

	suffix := time.Now().UnixNano()
	config := Config{
		Bucket:       fmt.Sprintf("testcase-%d", suffix),
		PublicBucket: fmt.Sprintf("testcase-public-%d", suffix),
	}
	if err := EnsureBuckets(context.Background(), config, "dev-library-checker-project"); err != nil {
		t.Fatal(err)
	}

	client, err := Connect(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	src := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(src, []byte("storage emulator content"), 0600); err != nil {
		t.Fatal(err)
	}

	const objectName = "nested/object.txt"
	if err := client.uploadFile(context.Background(), client.bucket, objectName, src); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), "downloaded.txt")
	if err := client.downloadToFile(context.Background(), client.bucket, objectName, dest); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(got)) != "storage emulator content" {
		t.Fatalf("downloaded content = %q", string(got))
	}
}
