package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/limingyao/excellent-go/config"
	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	Addr            string        `yaml:"addr"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifeTime time.Duration `yaml:"conn_max_life_time"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

func (c *Configuration) Init() error {
	return nil
}

func New(c *Configuration) *sqlx.DB {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn := "%s:%s@tcp(%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=Local&parseTime=true"
	conn = fmt.Sprintf(conn, c.User, c.Password, c.Addr, c.Database)
	db, err := sqlx.Open("mysql", conn)
	if err != nil {
		log.WithError(err).Fatalf("open mysql fail")
	}
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.ConnMaxLifeTime)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	if err := db.PingContext(ctx); err != nil {
		log.WithError(err).Fatalf("ping mysql fail")
	}

	return db
}

var configBytes = []byte(`
addr: ${DB_ADDR}
user: ${DB_USER}
password: ${DB_PASSWORD}
database: ${DB_DATABASE}
max_open_conns: 128
max_idle_conns: 128
conn_max_life_time: 500s
conn_max_idle_time: 120s
`)

func NewFromEnv() *sqlx.DB {
	cfg := &Configuration{}
	if err := config.Unmarshal(configBytes, cfg); err != nil {
		log.WithError(err).Fatal()
	}
	return New(cfg)
}
