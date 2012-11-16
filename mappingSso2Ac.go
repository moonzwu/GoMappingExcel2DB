package main

import (
	"GoUUID"
	_ "code.google.com/p/go-mysql-driver/mysql"
	"container/list"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"xlsx"
)

type MigrationRecord struct {
	name_cn    string
	mobile     string
	phone      string
	email      string
	account_id string
	sso_user   string
}

func findMatchRow(sheet *xlsx.Sheet, name string, index int, condition string) string {
	for _, row := range sheet.Rows {
		if row.Cells[1].String() == name && strings.Contains(row.Cells[index].String(), condition) {
			return row.Cells[0].String()
		}
	}

	return ""
}

func outputUnmatchedRow(sheet *xlsx.Sheet, name string) *xlsx.Row {
	for _, row := range sheet.Rows {
		if row.Cells[1].String() == name {
			return row
		}
	}

	return nil
}

func initMigrationRecords(db *sql.DB, recordList *list.List) {
	rows, e := db.Query("SELECT up.name_cn AS name_cn, up.mobile AS mobile, up.phone AS phone, ac.email AS email, up.`account_id` AS account_id FROM `user_profile` up , account ac WHERE up.`account_id` = ac.id AND up.expired_date IS NULL")
	if e != nil {
		return
	}

	for rows.Next() {
		var mr MigrationRecord
		rows.Scan(&mr.name_cn, &mr.mobile, &mr.phone, &mr.email, &mr.account_id)
		recordList.PushBack(mr)
	}
}

func updateDatabase(db *sql.DB, mr MigrationRecord) {
	uid, _ := uuid.NewV4()
	stmt, _ := db.Prepare("INSERT INTO sso2ac(`id`, `created_date`, `expired_date`, `old_id`, `updated_date`, `sso_user_name`, `account_id`) VALUES(?, now(), NULL, NULL, NULL, ?, ?)")
	tx, _ := db.Begin()
	tx.Stmt(stmt).Exec(uid.String(), mr.sso_user, mr.account_id)

	stmt1, _ := db.Prepare("UPDATE `user_profile` SET `sso_user_name` = ? WHERE `account_id` = ?")
	tx.Stmt(stmt1).Exec(uid.String(), mr.account_id)
	tx.Commit()

}

func main() {
	var xlFile *xlsx.File
	var err error

	if xlFile, err = xlsx.OpenFile("/Users/twer/Documents/BmwProject/dealerinfo.xlsx"); err != nil {
		fmt.Println(err)
		return
	}

	db, e := sql.Open("mysql", "root@/bmw?charset=utf8")
	defer db.Close()
	if e != nil {
		return
	}

	recordList := list.New()
	initMigrationRecords(db, recordList)

	fmt.Println("migration record number: ", recordList.Len())

	count := 0

	file, err := os.OpenFile("无法匹配到的用户.csv", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	b := []byte{0xEF, 0xBB, 0xBF}
	n, err := file.Write(b)
	if err != nil {
		panic(err)
	}

	fmt.Println("Write ", n, " bytes")

	csvHead := []string{"姓名", "电话", "邮箱", "备注"}
	w := csv.NewWriter(file)
	w.Write(csvHead)

	for e := recordList.Front(); e != nil; e = e.Next() {
		v := e.Value.(MigrationRecord)
		id := findMatchRow(xlFile.Sheets[0], v.name_cn, 12, v.email)

		f := func() string {
			if v.mobile == "" {
				return v.phone
			}
			return v.mobile
		}

		if id == "" {
			id = findMatchRow(xlFile.Sheets[0], v.name_cn, 10, f())
		}

		if id != "" {
			v.sso_user = id
			count++
			updateDatabase(db, v)
		} else {
			remark := ""
			r := outputUnmatchedRow(xlFile.Sheets[0], v.name_cn)
			if r == nil {
				remark = "Excel文件中没有该人信息"
			} else {
				remark = "与Excel文件中信息不匹配，excel中的信息为[" + r.Cells[1].String() + "," + r.Cells[10].String() + "," + r.Cells[12].String() + "]"
			}
			csvRecord := []string{v.name_cn, f(), v.email, remark}
			if err := w.Write(csvRecord); err != nil {
				panic(err)
			}

			fmt.Println(v, " ")
		}

	}
	w.Flush()

	fmt.Println("There are ", count, " record matched")

}
