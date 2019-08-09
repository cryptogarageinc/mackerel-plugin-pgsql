package mppgsql

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"

	// PostgreSQL Driver
	_ "github.com/lib/pq"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	"github.com/mackerelio/golib/logging"
)

var logger = logging.GetLogger("metrics.plugin.postgres")

// PgSQLPlugin mackerel plugin for PostgreSQL
type PgSQLPlugin struct {
	Host       string
	Port       string
	Username   string
	Password   string
	SSLmode    string
	Prefix     string
	Timeout    int
	Tempfile   string
	SQLConfigs []SQLConfig
	Option     string
}

// Config is collection of SQLConfig
type Config struct {
	SQLConfig []SQLConfig
}

// SQLConfig is struct of configuration for sql
type SQLConfig struct {
	Key          string
	Label        string
	MetricsName  string
	MetricsLabel string
	Unit         string
	SQL          string
}

func fetchSQL(db *sqlx.DB, s *SQLConfig) (map[string]interface{}, error) {
	rows, err := db.Query(s.SQL)
	if err != nil {
		logger.Errorf("Failed to select. %s", err)
		return nil, err
	}

	stat := map[string]interface{}{
		s.MetricsName: 0.0,
	}

	for rows.Next() {
		var count float64
		if err := rows.Scan(&count); err != nil {
			logger.Warningf("Failed to scan %s", err)
			continue
		}
		stat[s.MetricsName] = count
	}

	return stat, nil
}

func mergeStat(dst, src map[string]interface{}) {
	for k, v := range src {
		dst[k] = v
	}
}

// MetricKeyPrefix returns the metrics key prefix
func (p PgSQLPlugin) MetricKeyPrefix() string {
	if p.Prefix == "" {
		p.Prefix = "postgres"
	}
	return p.Prefix
}

// FetchMetrics interface for mackerelplugin
func (p PgSQLPlugin) FetchMetrics() (map[string]interface{}, error) {

	cmd := fmt.Sprintf("user=%s host=%s port=%s sslmode=%s connect_timeout=%d %s", p.Username, p.Host, p.Port, p.SSLmode, p.Timeout, p.Option)
	if p.Password != "" {
		cmd = fmt.Sprintf("password=%s %s", p.Password, cmd)
	}

	db, err := sqlx.Connect("postgres", cmd)
	if err != nil {
		logger.Errorf("FetchMetrics: %s", err)
		return nil, err
	}
	defer db.Close()

	stat := make(map[string]interface{})
	for _, s := range p.SQLConfigs {
		statCount, err := fetchSQL(db, &s)
		if err != nil {
			return nil, err
		}
		mergeStat(stat, statCount)
	}

	return stat, err
}

// GraphDefinition interface for mackerelplugin
func (p PgSQLPlugin) GraphDefinition() map[string]mp.Graphs {
	var graphdef = make(map[string]mp.Graphs)
	for _, s := range p.SQLConfigs {
		graphdef[s.Key] = mp.Graphs{
			Label: s.Label,
			Unit:  s.Unit,
			Metrics: []mp.Metrics{
				{Name: s.MetricsName, Label: s.MetricsLabel},
			},
		}
	}

	return graphdef
}

// Do the plugin
func Do() {
	optHost := flag.String("hostname", "localhost", "Hostname to login to")
	optPort := flag.String("port", "5432", "Database port")
	optUser := flag.String("user", "", "Postgres User")
	optDatabase := flag.String("database", "", "Database name")
	optPass := flag.String("password", os.Getenv("PGPASSWORD"), "Postgres Password")
	optPrefix := flag.String("metric-key-prefix", "postgres", "Metric key prefix")
	optSSLmode := flag.String("sslmode", "disable", "Whether or not to use SSL")
	optConnectTimeout := flag.Int("connect_timeout", 5, "Maximum wait for connection, in seconds.")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	optSQLConfig := flag.String("sqlconfig", "", "Sql config file")
	flag.Parse()

	if *optUser == "" {
		logger.Warningf("user is required")
		flag.PrintDefaults()
		os.Exit(1)
	}
	option := ""
	if *optDatabase != "" {
		option = fmt.Sprintf("dbname=%s", *optDatabase)
	}

	var pgsql PgSQLPlugin
	pgsql.Host = *optHost
	pgsql.Port = *optPort
	pgsql.Username = *optUser
	pgsql.Password = *optPass
	pgsql.Prefix = *optPrefix
	pgsql.SSLmode = *optSSLmode
	pgsql.Timeout = *optConnectTimeout
	pgsql.Option = option

	var config Config
	_, err := toml.DecodeFile(*optSQLConfig, &config)
	if err != nil {
		logger.Errorf("Failed to read sql config file. %s", err)
		flag.PrintDefaults()
		os.Exit(1)
	}
	pgsql.SQLConfigs = config.SQLConfig

	helper := mp.NewMackerelPlugin(pgsql)

	helper.Tempfile = *optTempfile
	helper.Run()
}
