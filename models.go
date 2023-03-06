package goscope2

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type Goscope2Log struct {
	ID  uint  `json:"id"`
	App int32 `gorm:"not null" json:"app"`
	// use generateMessageHash()
	Hash string `gorm:"index,not null" json:"hash"`
	// one of: `FATAL` `ERROR` `WARNING` `INFO`
	Severity string `gorm:"not null" json:"severity"`
	Message  string `gorm:"not null" json:"message"`

	Origin    string `json:"origin"`
	UserAgent string `json:"user_agent"`

	// javascript

	URL string `json:"url"`

	// http server

	Status int `json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

// first 10 chars of the md5 hash of message
func generateMessageHash(message string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(message)))[0:10]
}

func maybeCheckAndPurge(db *gorm.DB, maxRecords int) error {
	if rand.Intn(20) != 0 {
		return nil
	}

	return checkAndPurge(db, maxRecords)
}
func checkAndPurge(db *gorm.DB, maxRecords int) error {
	var count int
	err := db.Raw(`SELECT COUNT(*) FROM goscope2_logs`).Scan(&count).Error
	if err != nil {
		return err
	}

	if count > maxRecords {
		db.Exec(`
DELETE FROM goscope2_logs
WHERE id IN (
	SELECT id FROM goscope2_logs
	ORDER BY created_at ASC LIMIT 100
)
		`)
	}

	return nil
}

func getAll(db *gorm.DB) (*[]Goscope2Log, error) {
	list := &[]Goscope2Log{}
	if err := db.Raw(`SELECT * FROM goscope2_logs ORDER BY created_at DESC`).Scan(list).Error; err != nil {
		return nil, err
	}

	return list, nil
}
