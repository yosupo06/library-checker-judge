package main

import (
	"context"
	"database/sql"
	"testing"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
)

func TestHackList_UserDisplay(t *testing.T) {
	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		// Register users
		authClient.registerUID("token1", "uid1")
		authClient.registerUID("token2", "uid2")
		
		if err := database.RegisterUser(db, "testuser1", "uid1"); err != nil {
			t.Fatal(err)
		}
		if err := database.RegisterUser(db, "testuser2", "uid2"); err != nil {
			t.Fatal(err)
		}
		
		// Create problem
		if err := database.SaveProblem(db, database.Problem{
			Name:      "aplusb",
			Title:     "A + B",
			Timelimit: 2000,
		}); err != nil {
			t.Fatal(err)
		}
		
		// Create submissions
		subID1, err := database.SaveSubmission(db, database.Submission{
			ProblemName: "aplusb",
			UserName:    sql.NullString{Valid: true, String: "testuser1"},
			Source:      "source1",
		})
		if err != nil {
			t.Fatal(err)
		}
		
		subID2, err := database.SaveSubmission(db, database.Submission{
			ProblemName: "aplusb", 
			UserName:    sql.NullString{Valid: true, String: "testuser2"},
			Source:      "source2",
		})
		if err != nil {
			t.Fatal(err)
		}
		
		// Create hacks with different user scenarios
		// 1. Hack by testuser1
		_, err = database.SaveHack(db, database.Hack{
			SubmissionID: subID1,
			UserName:     sql.NullString{Valid: true, String: "testuser1"},
			TestCaseTxt:  []byte("test case 1"),
			Status:       "AC",
		})
		if err != nil {
			t.Fatal(err)
		}
		
		// 2. Hack by testuser2
		_, err = database.SaveHack(db, database.Hack{
			SubmissionID: subID2,
			UserName:     sql.NullString{Valid: true, String: "testuser2"},
			TestCaseTxt:  []byte("test case 2"),
			Status:       "WA",
		})
		if err != nil {
			t.Fatal(err)
		}
		
		// 3. Anonymous hack (UserName is null)
		_, err = database.SaveHack(db, database.Hack{
			SubmissionID: subID1,
			UserName:     sql.NullString{Valid: false, String: ""},
			TestCaseTxt:  []byte("anonymous test case"),
			Status:       "WJ",
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	
	// Test HackList API
	resp, err := client.HackList(context.Background(), &pb.HackListRequest{
		Skip:  0,
		Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	
	if len(resp.Hacks) != 3 {
		t.Fatalf("Expected 3 hacks, got %d", len(resp.Hacks))
	}
	
	// Verify user names are correctly displayed
	userNameCount := make(map[string]int)
	anonymousCount := 0
	
	for _, hack := range resp.Hacks {
		if hack.UserName != nil {
			userNameCount[*hack.UserName]++
			
			// Verify user names are valid
			if *hack.UserName != "testuser1" && *hack.UserName != "testuser2" {
				t.Fatalf("Unexpected user name: %s", *hack.UserName)
			}
		} else {
			anonymousCount++
		}
	}
	
	// Should have 1 hack each by testuser1 and testuser2, and 1 anonymous hack
	if userNameCount["testuser1"] != 1 {
		t.Fatalf("Expected 1 hack by testuser1, got %d", userNameCount["testuser1"])
	}
	if userNameCount["testuser2"] != 1 {
		t.Fatalf("Expected 1 hack by testuser2, got %d", userNameCount["testuser2"])
	}
	if anonymousCount != 1 {
		t.Fatalf("Expected 1 anonymous hack, got %d", anonymousCount)
	}
	
	// Test filtering by user
	resp, err = client.HackList(context.Background(), &pb.HackListRequest{
		Skip:  0,
		Limit: 10,
		User:  "testuser1",
	})
	if err != nil {
		t.Fatal(err)
	}
	
	if len(resp.Hacks) != 1 {
		t.Fatalf("Expected 1 hack by testuser1, got %d", len(resp.Hacks))
	}
	
	if resp.Hacks[0].UserName == nil || *resp.Hacks[0].UserName != "testuser1" {
		t.Fatalf("Expected hack by testuser1, got %v", resp.Hacks[0].UserName)
	}
}

func TestHackInfo_UserDisplay(t *testing.T) {
	var hackWithUserID, hackAnonymousID int32
	
	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		// Register user
		authClient.registerUID("token1", "uid1")
		
		if err := database.RegisterUser(db, "testuser", "uid1"); err != nil {
			t.Fatal(err)
		}
		
		// Create problem
		if err := database.SaveProblem(db, database.Problem{
			Name:      "aplusb",
			Title:     "A + B",
			Timelimit: 2000,
		}); err != nil {
			t.Fatal(err)
		}
		
		// Create submission
		subID, err := database.SaveSubmission(db, database.Submission{
			ProblemName: "aplusb",
			UserName:    sql.NullString{Valid: true, String: "testuser"},
			Source:      "source",
		})
		if err != nil {
			t.Fatal(err)
		}
		
		// Create hack with user
		hackWithUserID, err = database.SaveHack(db, database.Hack{
			SubmissionID: subID,
			UserName:     sql.NullString{Valid: true, String: "testuser"},
			TestCaseTxt:  []byte("test case"),
			Status:       "AC",
		})
		if err != nil {
			t.Fatal(err)
		}
		
		// Create anonymous hack
		hackAnonymousID, err = database.SaveHack(db, database.Hack{
			SubmissionID: subID,
			UserName:     sql.NullString{Valid: false, String: ""},
			TestCaseTxt:  []byte("anonymous test case"),
			Status:       "WJ", 
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	
	// Test HackInfo for hack with user
	resp, err := client.HackInfo(context.Background(), &pb.HackInfoRequest{
		Id: hackWithUserID,
	})
	if err != nil {
		t.Fatal(err)
	}
	
	if resp.Overview.UserName == nil {
		t.Fatal("Expected UserName to be set for hack with user")
	}
	if *resp.Overview.UserName != "testuser" {
		t.Fatalf("Expected UserName to be 'testuser', got '%s'", *resp.Overview.UserName)
	}
	
	// Test HackInfo for anonymous hack
	resp, err = client.HackInfo(context.Background(), &pb.HackInfoRequest{
		Id: hackAnonymousID,
	})
	if err != nil {
		t.Fatal(err)
	}
	
	if resp.Overview.UserName != nil {
		t.Fatalf("Expected UserName to be nil for anonymous hack, got '%s'", *resp.Overview.UserName)
	}
}

func TestHackAPI_UserCreation(t *testing.T) {
	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		// Register user
		authClient.registerUID("token1", "uid1")
		
		if err := database.RegisterUser(db, "testuser", "uid1"); err != nil {
			t.Fatal(err)
		}
		
		// Create problem
		if err := database.SaveProblem(db, database.Problem{
			Name:      "aplusb",
			Title:     "A + B",
			Timelimit: 2000,
		}); err != nil {
			t.Fatal(err)
		}
		
		// Create submission
		_, err := database.SaveSubmission(db, database.Submission{
			ProblemName: "aplusb",
			UserName:    sql.NullString{Valid: true, String: "testuser"},
			Source:      "source",
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	
	// Create hack as authenticated user
	hackResp, err := client.Hack(contextWithToken(context.Background(), "token1"), &pb.HackRequest{
		Submission: 1,
		TestCase:   &pb.HackRequest_Txt{Txt: []byte("test case")},
	})
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify the created hack has correct user information
	infoResp, err := client.HackInfo(context.Background(), &pb.HackInfoRequest{
		Id: hackResp.Id,
	})
	if err != nil {
		t.Fatal(err)
	}
	
	if infoResp.Overview.UserName == nil {
		t.Fatal("Expected UserName to be set for hack created by authenticated user")
	}
	if *infoResp.Overview.UserName != "testuser" {
		t.Fatalf("Expected UserName to be 'testuser', got '%s'", *infoResp.Overview.UserName)
	}
	
	// Create hack as anonymous user (no token)
	hackResp, err = client.Hack(context.Background(), &pb.HackRequest{
		Submission: 1,
		TestCase:   &pb.HackRequest_Txt{Txt: []byte("anonymous test case")},
	})
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify the created hack is anonymous
	infoResp, err = client.HackInfo(context.Background(), &pb.HackInfoRequest{
		Id: hackResp.Id,
	})
	if err != nil {
		t.Fatal(err)
	}
	
	if infoResp.Overview.UserName != nil {
		t.Fatalf("Expected UserName to be nil for anonymous hack, got '%s'", *infoResp.Overview.UserName)
	}
}