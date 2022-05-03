package meta

import (
	mydb "filestore-server/db"
)

//FileMeta: 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var FileMetas map[string]FileMeta

func init() {
	FileMetas = make(map[string]FileMeta)
}

//UpdateFileMeta:新增 / 更新文件元信息
func UpdateFileMeta(fmeta FileMeta) {
	FileMetas[fmeta.FileSha1] = fmeta
}

//UpdateFileMetaDB: 新增 / 更新文件元信息到mysqlzh中
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return mydb.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

//GetFileMeta: 通过文件 sha1 获取文件元信息
func GetFileMeta(fileSha1 string) FileMeta {
	return FileMetas[fileSha1]
}

// GetFileMetaDB: 从Mysql中获取文件元信息
func GetFileMetaDB(fileSha1 string) (FileMeta, error) {
	tfile, err := mydb.GetFileMeta(fileSha1)
	if err != nil {
		return FileMeta{}, err
	}
	fmeta := FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return fmeta, nil
}

// 删除元信息
func RemoveFileMeta(fileSha1 string) {
	delete(FileMetas, fileSha1)
}
