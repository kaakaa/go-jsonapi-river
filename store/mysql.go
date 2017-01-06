package store

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type DataSource struct {
	DB *sql.DB
}

func NewDB() DataSource {
	db, err := sql.Open("mysql", "root:root@tcp(10.25.165.168:13306)/redash_sample")
	if err != nil {
		panic(err)
	}

	return DataSource{DB: db}
}

func (d *DataSource) Write(tableName string, sql string, data Data) {
	stmtIns, err := d.DB.Prepare(fmt.Sprintf(sql, tableName))
	if err != nil {
		panic(err)
	}
	defer stmtIns.Close()

	_, err = stmtIns.Exec(data.Name, data.Date, data.LoC)
}

func (d *DataSource) WriteSQL(sql string) {
	_, err := d.DB.Exec(sql)
	if err != nil {
		panic(err)
	}
}

type Data struct {
	Name string
	Date string
	LoC  int
}
