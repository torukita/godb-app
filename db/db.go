package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

const(
	DBName = "sampledb"
	DBTable = "sample_data"
)

const schemaPostgres = `
CREATE TABLE IF NOT EXISTS %s (
    id SERIAL NOT NULL,
    memo varchar(255) NOT NULL,
    sub_memo varchar(255) NOT NULL DEFAULT '',
    timestamp timestamp default CURRENT_TIMESTAMP,
    PRIMARY KEY(id)
);`

const schemaMysql = `
CREATE TABLE IF NOT EXISTS %s (
    id INT NOT NULL AUTO_INCREMENT,
    memo varchar(255) NOT NULL,
    sub_memo varchar(255) NOT NULL DEFAULT '',
    timestamp timestamp default CURRENT_TIMESTAMP,
    PRIMARY KEY(id)
);`

var (
	config map[string]DBConf = make(map[string]DBConf)
	environment = "development"
	dbx *sqlx.DB = &sqlx.DB{}
)

type DBConf struct {
	Driver string `yaml:"driver"`
	Dsn    string `yaml:"open"`
}

func Load(file string) error {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return err
	}
	return nil
}

func SetEnv(name string) error {
	if _, ok := config[name]; ok {
		environment = name
		return nil
	}
	return fmt.Errorf("not found environment")
}

func Connect() error {
	driver := config[environment].Driver
	dsn := config[environment].Dsn

	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return err
	}
	dbx = db
	return nil
}

func Close() {
	dbx.Close()
}

type Memo struct {
	ID        int       `db:"id"`
	Memo      string    `db:"memo"`
	SubMemo   string    `db:"sub_memo"`
	Timestamp time.Time `db:"timestamp"`
}

func (m Memo) Dump() {
	fmt.Printf("%+v\n", m)
	// fmt.Printf("id:%d memo:%s sub_memo:%s timestamp:%s\n",
	//	m.ID, m.Memo, m.SubMemo, m.Timestamp)
}

func DumpMemo() error {
	sql := fmt.Sprintf("SELECT * from %s", DBTable)
	rows, err := dbx.Queryx(sql)
	if err != nil {
		return err
	}
	num := 0
	var memos []Memo
	for rows.Next() {
		var memo Memo
		err := rows.StructScan(&memo)
		if err != nil {
			return err
		}
		memos = append(memos, memo)
		num++
	}

	for _, m := range memos {
		m.Dump()
	}
	fmt.Printf("Total Count=%d\n", num)
	return nil
}

func AddMemo(memo, sub_memo string) error {
	sql := fmt.Sprintf("INSERT INTO %s (memo, sub_memo) VALUES (?, ?)", DBTable)
	_, err := dbx.Exec(dbx.Rebind(sql), memo, sub_memo)
	if err != nil {
		return err
	}
	return nil
}

func DeleteMemos() error {
	sql := fmt.Sprintf("TRUNCATE TABLE %s", DBTable)
	_, err := dbx.Exec(sql)
	return err
}

func CountMemo() (count int) {
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s", DBTable)
	row := dbx.QueryRowx(sql)
	row.Scan(&count)
	return
}

func CreateTable() error {
	var sql string
	driver := config[environment].Driver
	if driver == "mysql" {
		sql = fmt.Sprintf(schemaMysql, DBTable)
	} else if driver == "postgres" {
		sql = fmt.Sprintf(schemaPostgres, DBTable)
	}
	_, err := dbx.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func DropTable() error {
	sql := fmt.Sprintf("DROP TABLE %s", DBTable)
	_, err := dbx.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}
