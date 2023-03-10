package goscope2

import (
	"crypto/md5"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	TYPE_LOG  = "log"
	TYPE_HTTP = "http"
	TYPE_JS   = "js"
)

type Goscope2Log struct {
	ID   uint   `json:"id"`
	Type string `gorm:"not null" json:"type" binding:"required,oneof='http' 'js' 'log'`
	// use generateMessageHash()
	Hash     string `gorm:"index,not null" json:"hash"`
	Severity string `gorm:"not null" json:"severity"  binding:"required,oneof='FATAL' 'ERROR' 'WARNING' 'INFO'`
	Message  string `gorm:"not null" json:"message" binding:"required,min=1"`

	Origin    string `json:"origin"`
	UserAgent string `json:"user_agent"`

	// javascript

	URL string `json:"url"`

	// http

	Status int `json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

// first 10 chars of the md5 hash of message
func (g *Goscope2Log) GenerateHash() {
	s := g.Message + g.Severity + g.Type
	g.Hash = fmt.Sprintf("%x", md5.Sum([]byte(s)))[0:10]
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
WHERE id NOT IN (
	SELECT id FROM goscope2_logs
	ORDER BY created_at DESC LIMIT ?
)
		`, maxRecords)
	}

	return nil
}

func getSome(db *gorm.DB, page int, ftype string) (*[]Goscope2Log, error) {
	list := &[]Goscope2Log{}
	if err := db.Debug().Raw(`
SELECT * FROM goscope2_logs
WHERE type = ?
ORDER BY created_at ASC
LIMIT 100 OFFSET ?
	`, ftype, (page-1)*100).Scan(list).Error; err != nil {
		return nil, err
	}

	return list, nil
}
