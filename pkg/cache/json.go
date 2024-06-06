package cache

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	_ "github.com/glebarez/go-sqlite"
	"log"
	"reflect"
	"regexp"
	"time"
)

type JsonCache struct {
	db    *sql.DB
	table string
}

func (c *JsonCache) init() error {
	_, err := c.db.Exec(`
		CREATE TABLE IF NOT EXISTS ` + c.table + ` (
			id INTEGER NOT NULL PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			time DATETIME NOT NULL,
			value TEXT
		);
	`)
	return err
}

func (c *JsonCache) Get(key string, value interface{}) bool {
	if reflect.ValueOf(value).Kind() != reflect.Ptr {
		panic("JsonCache:Get value must be pointer")
	}

	query := "SELECT value FROM " + c.table + " WHERE key = ?"
	row := c.db.QueryRow(query, key)
	if row == nil {
		return false
	}

	var valueStr string
	err := row.Scan(&valueStr)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("ERROR: JsonCache(GET:%s:%s): %v", c.table, key, err)
		}
		return false
	}

	err = json.Unmarshal([]byte(valueStr), value)
	if err != nil {
		log.Printf("ERROR: JsonCache(GET:%s:%s): %v", c.table, key, err)
		return false
	}

	// Update `time` on successful access
	_, _ = c.db.Exec("UPDATE "+c.table+" SET time = datetime('now') WHERE key = ?", key)

	return true
}

func (c *JsonCache) Set(key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		log.Printf("ERROR: JsonCache(SET:%s:%s): %v", c.table, key, err)
		return err
	}

	_ = c.Delete(key)
	_, err = c.db.Exec("INSERT INTO "+c.table+" (key, time, value) VALUES (?, datetime('now'), ?)", key, string(valueBytes))
	if err != nil {
		log.Printf("ERROR: JsonCache(SET:%s:%s): %v", c.table, key, err)
	}

	return err
}

func (c *JsonCache) Delete(key string) error {
	_, err := c.db.Exec("DELETE FROM "+c.table+" WHERE key = ?", key)
	if err != nil {
		log.Printf("ERROR: JsonCache(DELETE:%s:%s): %v", c.table, key, err)
	}

	return err
}

func (c *JsonCache) Clean(maxAge time.Duration) error {
	_, err := c.db.Exec("DELETE FROM "+c.table+" WHERE time < ?", time.Now().Add(-maxAge))
	if err != nil {
		log.Printf("ERROR: JsonCache(CLEAN:%s): %v", c.table, err)
	}

	return err

}

func NewJsonCache(cfg config.CacheConfig) (ObjectCache, error) {
	tableRe := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !tableRe.MatchString(cfg.Table) {
		return nil, errors.New("invalid table name")
	}

	// Attempt to create a sqlite db file
	db, err := sql.Open("sqlite", cfg.File)
	if err != nil {
		return nil, err
	}

	c := &JsonCache{
		db:    db,
		table: cfg.Table,
	}

	return c, c.init()
}
