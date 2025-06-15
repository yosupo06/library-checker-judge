package main

import (
	"context"
	"strings"
	"testing"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
)

func TestSubmit(t *testing.T) {
	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		if err := database.SaveProblem(db, DUMMY_PROBLEM); err != nil {
			t.Fatal("Failed to save problem:", err)
		}
	})

	ctx := context.Background()
	_, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source:  "dummy-src",
		Lang:    "cpp",
	})
	if err != nil {
		t.Fatal("Failed to submit", err)
	}
	t.Log(err)
}

func TestSubmissionSortOrderList(t *testing.T) {
	client := createTestAPIClient(t)

	ctx := context.Background()
	for _, order := range []string{"", "-id", "+time"} {
		_, err := client.SubmissionList(ctx, &pb.SubmissionListRequest{
			Skip:  0,
			Limit: 100,
			Order: order,
		})
		if err != nil {
			t.Fatal("Failed SubmissionList Order:", order)
		}
	}
	_, err := client.SubmissionList(ctx, &pb.SubmissionListRequest{
		Skip:  0,
		Limit: 100,
		Order: "dummy",
	})
	if err == nil {
		t.Fatal("Success SubmissionList Dummy Order")
	}
	t.Log(err)
}

func TestSubmitBig(t *testing.T) {
	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		if err := database.SaveProblem(db, DUMMY_PROBLEM); err != nil {
			t.Fatal("Failed to save problem:", err)
		}
	})

	ctx := context.Background()
	bigSrc := strings.Repeat("a", 3*1000*1000) // 3 MB
	_, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source:  bigSrc,
		Lang:    "cpp",
	})
	if err == nil {
		t.Fatal("Success to submit big source")
	}
	t.Log(err)
}

func TestSubmitUnknownLang(t *testing.T) {
	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		if err := database.SaveProblem(db, DUMMY_PROBLEM); err != nil {
			t.Fatal("Failed to save problem:", err)
		}
	})

	ctx := context.Background()
	for _, lang := range []string{"invalid-lang", "checker"} {
		_, err := client.Submit(ctx, &pb.SubmitRequest{
			Problem: "aplusb",
			Source:  "dummy-src",
			Lang:    lang,
		})
		if err == nil {
			t.Fatal("Success to submit unknown language", err)
		}
		t.Log(err)
	}
}

func TestAnonymousRejudge(t *testing.T) {
	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		if err := database.SaveProblem(db, DUMMY_PROBLEM); err != nil {
			t.Fatal("Failed to save problem:", err)
		}
	})

	ctx := context.Background()
	src := strings.Repeat("a", 1000)
	resp, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: DUMMY_PROBLEM.Name,
		Source:  src,
		Lang:    "cpp",
	})
	if err != nil {
		t.Fatal("Unsuccess to submit source:", err)
	}
	_, err = client.Rejudge(ctx, &pb.RejudgeRequest{
		Id: resp.Id,
	})
	if err == nil {
		t.Fatal("Success to rejudge")
	}
}
