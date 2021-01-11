module github.com/kyleconroy/sqlparse/internal/integration

replace github.com/kyleconroy/sqlparse => ../../

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/kyleconroy/sqlparse v0.0.0-20210110220917-d7b13c2af3d8
)

go 1.13
