package mppgsql

import (
	"flag"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"

	// PostgreSQL Driver
	_ "github.com/lib/pq"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	"github.com/mackerelio/golib/logging"
)

var logger = logging.GetLogger("metrics.plugin.postgres")

// PgSQLPlugin mackerel plugin for PostgreSQL
type PgSQLPlugin struct {
	Host         string
	Port         string
	Username     string
	Password     string
	SSLmode      string
	Prefix       string
	Timeout      int
	Tempfile     string
	Option       string
	Column       string
	Table        string
	Condition    string
	Key          string
	Label        string
	MetricsName  string
	MetricsLabel string
	Unit         string
}

func fetchSQL(db *sqlx.DB, p *PgSQLPlugin) (map[string]interface{}, error) {
	sql := fmt.Sprintf("SELECT %s FROM %s", p.Column, p.Table)
	if len(p.Condition) > 0 {
		sql += fmt.Sprintf(" WHERE %s", p.Condition)
	}
	rows, err := db.Query(sql)
	if err != nil {
		logger.Errorf("Failed to select. %s", err)
		return nil, err
	}

	stat := map[string]interface{}{
		p.MetricsName: 0.0,
	}

	for rows.Next() {
		result := float64(0)
		rows.Scan(&result)
		stat[p.MetricsName] = result
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
	statCount, err := fetchSQL(db, &p)
	if err != nil {
		return nil, err
	}
	mergeStat(stat, statCount)

	return stat, err
}

// GraphDefinition interface for mackerelplugin
func (p PgSQLPlugin) GraphDefinition() map[string]mp.Graphs {
	var graphdef = map[string]mp.Graphs{
		p.Key: {
			Label: p.Label,
			Unit:  p.Unit,
			Metrics: []mp.Metrics{
				{Name: p.MetricsName, Label: p.MetricsLabel},
			},
		},
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
	optColumn := flag.String("column", "", "column of select statement")
	optTable := flag.String("table", "", "table of select statement")
	optCondition := flag.String("condition", "", "where clause of select statement")
	optKey := flag.String("key", "", "graph key")
	optLabel := flag.String("label", "", "graph label")
	optMetricsName := flag.String("metricsname", "", "graph metrics name")
	optMetricsLabel := flag.String("metricslabel", "", "graph mertics label")
	optUnit := flag.String("unit", "integer", "graph unit")
	flag.Parse()

	if *optUser == "" {
		logger.Warningf("user is required")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *optColumn == "" {
		logger.Warningf("column is required")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *optTable == "" {
		logger.Warningf("table is required")
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
	pgsql.Column = *optColumn
	pgsql.Table = *optTable
	pgsql.Condition = *optCondition
	pgsql.Key = *optKey
	pgsql.Label = *optLabel
	pgsql.MetricsName = *optMetricsName
	pgsql.MetricsLabel = *optMetricsLabel
	pgsql.Unit = *optUnit

	helper := mp.NewMackerelPlugin(pgsql)

	helper.Tempfile = *optTempfile
	helper.Run()
}
