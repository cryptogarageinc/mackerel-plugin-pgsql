# mackerel-plugin-pgsql
mackerel plugin for executing sql on postgres

# How to use
ex) `$ mackerel-plugin-pgsql -user auth -database auth -hostname 127.0.0.1 -password password -port 5432 -column "COUNT(1)" -table "SAMPLE" -condition "status = 'active'" -key key -label label -metricsname count -metricslabel count -unit integer`

```
  -column string
        column of select statement
  -condition string
        where clause of select statement
  -connect_timeout int
        Maximum wait for connection, in seconds. (default 5)
  -database string
        Database name
  -hostname string
        Hostname to login to (default "localhost")
  -key string
        graph key
  -label string
        graph label
  -merticslabel string
        graph mertics label
  -metric-key-prefix string
        Metric key prefix (default "postgres")
  -metricsname string
        graph metrics name
  -password string
        Postgres Password
  -port string
        Database port (default "5432")
  -sslmode string
        Whether or not to use SSL (default "disable")
  -table string
        table of select statement
  -tempfile string
        Temp file name
  -unit string
        graph unit (default "integer")
  -user string
        Postgres User
```

# How to release for mkr install
1. `$ make setup`
1. `$ git tag v0.19.1` (タグ名は適宜置き換えること)
1. `$ GITHUB_TOKEN=... script/release.sh` (GITHUB_TOKENはあらかじめ発行しておくこと)
