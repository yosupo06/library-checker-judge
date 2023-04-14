package main

import (
	"errors"
	"log"
	"time"

	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RegistrationStatus : status of judge registration
type RegistrationStatus int

const (
	// Waiting : Waiting Judge
	Waiting RegistrationStatus = iota
	// JudgingBySelf : ジャッジ中
	JudgingBySelf
	// JudgingByOther : 他がジャッジ中
	JudgingByOther
	// Finished : ジャッジ終了
	Finished
)

func (status RegistrationStatus) String() string {
	switch status {
	case Waiting:
		return "Waiting"
	case JudgingBySelf:
		return "JudgingBySelf"
	case JudgingByOther:
		return "JudgingByOther"
	case Finished:
		return "Finished"
	default:
		return "Unknown"
	}
}

func currentRegistrationStatus(sub *database.Submission, judgeName string) RegistrationStatus {
	now := time.Now()

	if sub.JudgeName != "" && sub.JudgePing.After(now) {
		if sub.JudgeName == judgeName {
			return JudgingBySelf
		}
		return JudgingByOther
	}

	if sub.JudgeName != "" {
		return Waiting
	}

	return Finished
}

func changeRegistrationStatus(db *gorm.DB, id int32, judgeName string, updateJudgeName string, expiration time.Duration, expect RegistrationStatus) error {
	return db.Transaction(func(tx *gorm.DB) error {
		sub := &database.Submission{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(sub, id).Error; err != nil {
			log.Print(err)
			return errors.New("Submission fetch failed")
		}

		status := currentRegistrationStatus(sub, judgeName)

		if status != expect {
			log.Printf("Expect(%v) != Actual(%v)", expect, status)
			return errors.New("actual status does not matched to expected status")
		}

		if err := tx.Model(&sub).Updates(map[string]interface{}{
			"judge_name": updateJudgeName,
			"judge_ping": time.Now().Add(expiration),
		}).Error; err != nil {
			log.Print(err)
			return errors.New("Submission update failed")
		}
		return nil
	})
}

func registerSubmission(db *gorm.DB, id int32, judgeName string, expiration time.Duration, expect RegistrationStatus) error {
	return changeRegistrationStatus(db, id, judgeName, judgeName, expiration, expect)
}

func updateSubmissionRegistration(db *gorm.DB, id int32, judgeName string, expiration time.Duration) error {
	return changeRegistrationStatus(db, id, judgeName, judgeName, expiration, JudgingBySelf)
}

func releaseSubmissionRegistration(db *gorm.DB, id int32, judgeName string) error {
	return changeRegistrationStatus(db, id, judgeName, "", -time.Second, JudgingBySelf)
}

func toWaitingJudge(db *gorm.DB, id int32, priority int32, after time.Duration) error {
	if err := registerSubmission(db, id, "#WaitingJudge", -time.Second, Finished); err != nil {
		return err
	}

	sub := &database.Submission{}
	if err := db.Take(sub, id).Error; err != nil {
		log.Print(err)
		return errors.New("failed to fetch submission")
	}
	sub.PrevStatus = sub.Status
	sub.Status = "WJ"
	if err := db.Save(sub).Error; err != nil {
		log.Print(err)
		return errors.New("failed to update status")
	}

	if err := PushTask(db, database.Task{
		Submission: id,
		Available:  time.Now().Add(after),
		Priority:   priority,
	}); err != nil {
		log.Print(err)
		return errors.New("cannot insert into queue")
	}

	return nil
}
