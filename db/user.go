package db

import (
	"encoding/json"
	mydb "filestore-server/db/mysql"

	"fmt"
)

// UserSignup:通过用户名及密码完成user表的注册操作
func UserSignup(username string, passwd string) bool {
	stmt, err := mydb.DBconn().Prepare("insert ignore into tbl_user(`user_name`,`user_pwd`)values(?,?)")
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, passwd)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	if rowsAffected, err := ret.RowsAffected(); nil == err && rowsAffected > 0 {
		return true
	}
	return false
}

//UserSignin : 校验用户名和密码
func UserSignin(username string, encPassword string) bool {
	stmt, err := mydb.DBconn().Prepare("select * from tbl_user where user_name = ? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("username not found:" + username)
		return false
	}

	pRows := mydb.ParseRows(rows)
	data, err := json.Marshal(pRows)

	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encPassword {
		if err != nil {
			fmt.Println(data)
		}
		return true
	}

	return false
}

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBconn().Prepare(
		"select user_name, signup_at from tbl_user where user_name = ? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	// 查询用户信息
	stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		return user, err
	}
	return user, nil
}
