package mybmwscript

import (
	_ "code.google.com/p/go-mysql-driver/mysql"
	"database/sql"
	"github.com/bmizerany/assert"
	"testing"
)

func TestConnectMySql(t *testing.T) {
	db, e := sql.Open("mysql", "root@/bmw?charset=utf8")
	defer db.Close()
	if e != nil {
		t.Error(e.Error())
	}
}

func TestGetMigrationTableRecrodCounts(t *testing.T) {
	db, e := sql.Open("mysql", "root@/bmw?charset=utf8")
	defer db.Close()

	if e != nil {
		t.Error(e.Error())
	}

	rows, e := db.Query("select * from migration_table")
	if e != nil {
		t.Error("query failed!")
	}

	count := 1
	for rows.Next() {
		count++
	}

	assert.Equal(t, 365, count)

}

func TestGetTheFirstLineOfMigrationRecrod(t *testing.T) {
	db, e := sql.Open("mysql", "root@/bmw?charset=utf8")
	defer db.Close()

	if e != nil {
		t.Error(e.Error())
	}

	rows, e := db.Query("select * from migration_table")
	if e != nil {
		t.Error("query failed!")
	}

	var record MigrationRecord
	rows.Next()
	rows.Scan(&record.name_cn, &record.mobile, &record.email, &record.account_id, &record.sso_user)

	expectedRecord := MigrationRecord{"谢曦", "", "cici.xie@gzbaoze.bmw.com.cn", "0000000036c59fdf0136c8f75ac701f8", ""}

	assert.Equal(t, record, expectedRecord)
}
