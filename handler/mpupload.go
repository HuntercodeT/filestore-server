package handler

import (
	rPool "filestore-server/cache"
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

// MultipartUploadInfo:初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// 初始化分块上传
func InitiialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invaild", nil).JSONBytes())
		return
	}

	// 2. 获得 redis 的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 生成分块上传的初始化信息
	upinfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().In(cstSh).UnixNano()),
		ChunkSize:  5 * 1024 * 1024, //5M
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}
	// 4. 将初始化信息写入到 redia 缓存
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "chunkcount", upinfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filehash", upinfo.FileHash)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filesize", upinfo.FileSize)
	rConn.Do("HMSET")
	// 5.  将初始化信息返回给客户端
	w.Write(util.NewRespMsg(0, "OK", upinfo).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	// username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkindex := r.Form.Get("index")

	// 2. 获得 redis 连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 获得文件句柄，用于存储分块内容
	fpath := "/tmp/filestore/" + uploadID + "/" + chunkindex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create("/tmp/filestore/" + uploadID + "/" + chunkindex)
	// todo : 每个分块带一个hash发送请求，服务端校验决定是否走秒传接口
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Chdir()
	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 4. 更新 redis 缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chunkidx_"+chunkindex, 1)
	// 5.返回处理结果给客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invaild", nil).JSONBytes())
		return
	}
	filename := r.Form.Get("filename")

	// 2. 获得 redis 连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 通过 uploadid 查询 redis 并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadID))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "conplete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	ChunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(k)
		} else if strings.HasPrefix(k, "chunkidx_") && v == "1" {
			ChunkCount++
		}
	}
	if totalCount != ChunkCount {
		w.Write(util.NewRespMsg(-2, "invaild request", nil).JSONBytes())
		return
	}
	// 4. TODO:合并分块
	// 5. 更新唯一文件表及用户文件表
	dblayer.OnFileUploadFinished(filehash, filename, int64(filesize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))

	// 6.响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
