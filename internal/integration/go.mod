module github.com/kyleconroy/sqlparse/internal/integration

replace github.com/kyleconroy/sqlparse => ../../

require (
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548
	github.com/cznic/parser v0.0.0-20160622100904-31edd927e5b1
	github.com/cznic/sortutil v0.0.0-20181122101858-f5f958428db8
	github.com/cznic/strutil v0.0.0-20171016134553-529a34b1c186
	github.com/cznic/y v0.0.0-20170802143616-045f81c6662a
	github.com/go-sql-driver/mysql v1.5.0
	github.com/kyleconroy/sqlparse v0.0.0-20210110220917-d7b13c2af3d8
	github.com/pingcap/errors v0.11.5-0.20201029093017-5a7df2af2ac7
)

go 1.13
