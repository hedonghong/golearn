package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var tmpDir string
var httpUrl string

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
}

func main() {
	flag.StringVar(&tmpDir, "dir", "", "存储临时语言文件目录，eg: -dir /tmp")
	flag.Parse()
	if len(tmpDir) <= 0 {
		panic(errors.New("请输入存储临时语言文件目录"))
	}
	fmt.Println(tmpDir)
	if !IsWriteDir(tmpDir) {
		panic(errors.New(tmpDir + "存储临时语言文件目录无法写入"))
	}
	if len(httpUrl) > 0 {
		http.HandleFunc("/upload", paddlespeech)
		http.HandleFunc("/", HelloServer)
		err := http.ListenAndServe(":12345", nil)
		fmt.Println(err)
	}
}

func IsWriteDir(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return info.IsDir() && unix.Access(path, unix.W_OK) == nil
	}
	return false
}

func FileIsExisted(filename string) bool {
	existed := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		existed = false
	}
	return existed
}

func uploadFile(r *http.Request) (string, *os.File, error) {
	uploadFile, _, err := r.FormFile("file")
	fileName := strconv.FormatInt(time.Now().UnixNano(), 10)
	// /Users/sky.he/code/python
	path := tmpDir + "/" + fileName + ".wav"
	file, err := os.Create(path)
	if err != nil {
		return "", nil, err
	}
	_, err = io.Copy(file, uploadFile)
	if err != nil {
		return "", nil, err
	}
	return path, file, nil
}

func downloadFile(r *http.Request) (string, *os.File, error) {
	fileUrl := r.URL.Query().Get("fileUrl")
	if len(fileUrl) <= 0 {
		return "", nil, errors.New("无法下载" + fileUrl)
	}
	ext := filepath.Ext(fileUrl)
	if ext != ".wav" {
		return "", nil, errors.New("目前只能解析wav后续的语言文件")
	}
	resp, err := http.Get(fileUrl)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	fileName := strconv.FormatInt(time.Now().UnixNano(), 10)
	// /Users/sky.he/code/python
	path := tmpDir + "/" + fileName + ".wav"
	file, err := os.Create(path)
	if err != nil {
		return "", nil, err
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", nil, err
	}
	return path, file, nil
}

func paddlespeech(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		io.WriteString(w, "必须POST请求")
		return
	}
	// 上传form-data文件 字段名称为file，上传服务器后解析返回
	//path, file, err := uploadFile(r)
	// 根据url上面的字段fileUrl传入文件网络路径，服务器下载后解析返回
	path, file, err := downloadFile(r)
	defer os.Remove(path)
	defer file.Close()

	agrs := []string{
		"asr",
		"--input",
		//"/Users/sky.he/code/python/test4.wav",
		path,
	}
	if !FileIsExisted(path) {
		io.WriteString(w, "文件不存在")
		return
	}
	cmd := exec.Command("paddlespeech", agrs...)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err = cmd.Run()
	if err != nil {
		io.WriteString(w, fmt.Sprintf("%s %s", stdErr.String(), err.Error()))
		return
	}
	fmt.Println(stdOut.String())
	fmt.Println(stdErr.String())
	outputStr := stdErr.String()
	reg, err := regexp.Compile("\\[    INFO\\] \\- ASR Result\\:(.*)")
	if err != nil {
		io.WriteString(w, fmt.Sprintf("%s", err.Error()))
		return
	}
	regStr := reg.FindStringSubmatch(outputStr)
	if len(regStr) == 2 {
		io.WriteString(w, regStr[1])
		return
	}
	io.WriteString(w, "抱歉！分析不出来～")
	return
}
