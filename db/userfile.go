package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
	"time"
)

// 用户文件结构
type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

var cstSh, _ = time.LoadLocation("Asia/Shanghai")

// 更新用户文件表
func OnUserFileUploadFinished(username, filehash, filename string, filesize int64) bool {
	stmt, err := mydb.DBconn().Prepare(
		"insert ignore into tbl_user_file(`user_name`,`file_sha1`,`file_size`," +
			"`file_name`,`upload_at`)values(?,?,?,?,?)")
	if err != nil {
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash, filesize, filename, time.Now().In(cstSh).Format("2006-01-02 15:04:05"))
	if err != nil {
		return false
	}
	return true
}

// QueryUserFileMetas: 批量获取用户信息
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mydb.DBconn().Prepare(
		"select file_sha1,file_size,file_name,upload_at," +
			"last_update from tbl_user_file where user_name=? limit ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		return nil, err
	}
	var userFiles []UserFile
	for rows.Next() {
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash, &ufile.FileSize, &ufile.FileName, &ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles, ufile)

	}
	return userFiles, nil
}
