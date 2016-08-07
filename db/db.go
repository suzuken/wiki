package db

import (
	"database/sql"
	"io"
	"io/ioutil"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v1"
)

// Configsは環境ごとの設定情報をもつ
type Configs map[string]*Config

// Openは指定された環境についてDBに接続します。
func (cs Configs) Open(env string) (*sql.DB, error) {
	config, ok := cs[env]
	if !ok {
		return nil, nil
	}
	return config.Open()
}

// Configはsql-migrateの設定ファイルと同じ形式を想定している
type Config struct {
	Datasource string `yaml:"datasource"`
}

// DSNは設定されているDSNを返します
func (c *Config) DSN() string {
	return c.Datasource
}

// OpenはConfigで指定されている接続先に接続する。
// MySQL固定
func (c *Config) Open() (*sql.DB, error) {
	return sql.Open("mysql", c.DSN())
}

// NewConfigsFromFileはConfigから設定を読み取る
func NewConfigsFromFile(path string) (Configs, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewConfigs(f)
}

// NewConfigsはio.ReaderからDB用設定を読み取る
func NewConfigs(r io.Reader) (Configs, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var configs Configs
	if err = yaml.Unmarshal(b, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}
