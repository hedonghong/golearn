package nethttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

var portTest string = ":20013"
var portRequest string = "http://127.0.0.1:20013"


func HttpMyServer(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)
	fmt.Println(r.Header)
	fmt.Println(r.URL.Query())
	fmt.Println(r.URL.Query().Get("test"))
	urlRequest := r.URL.Query()
	urlRequest.Add("test1", "222")
	fmt.Println(urlRequest)
	// post form内容 发送请求使用r.PostForm()
	fmt.Println(r.PostFormValue("test"))
	bodyContent, err := ioutil.ReadAll(r.Body)
	if err != nil && err != io.EOF {
		fmt.Println(err.Error())
	}
	time.Sleep(time.Second * 5)
	fmt.Println(string(bodyContent))
}

func TestHttpServer(t *testing.T) {
	http.HandleFunc("/index", HttpMyServer)
	http.ListenAndServe(portTest, nil)
}

func TestGet(t *testing.T) {
	resp, _ := http.Get(portRequest+"/index")
	defer resp.Body.Close()
	bodyContent, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(bodyContent))
}

func TestPost(t *testing.T) {
	param := url.Values{}
	param.Add("test", "TestPost")
	resp, _ := http.PostForm(portRequest+"/index", param)
	defer resp.Body.Close()
	bodyContent, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(bodyContent))
}


func TestGetParam(t *testing.T) {
	param := url.Values{}
	param.Add("test", "TestGetParam")

	newUrl, _ := url.Parse(portRequest+"/index")
	newUrl.RawQuery = param.Encode()

	resp, _ := http.Get(newUrl.String())
	defer resp.Body.Close()
	bodyContent, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(bodyContent))
}

func TestGetHeader(t *testing.T) {
	client := &http.Client{}
	requestGet, _ := http.NewRequest("GET", portRequest+"/index", nil)

	requestGet.Header.Add("header", "sky.he")

	resp, _ := client.Do(requestGet)
	defer resp.Body.Close()
	bodyContent, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(bodyContent))
}

func TestPostContent1(t *testing.T) {
	client := &http.Client{}

	data := make(map[string]interface{})
	data["test"] = "TestPostContent1"
	jsonData, _ := json.Marshal(data)
	requetPost, _ := http.NewRequest("POST", portRequest+"/index", bytes.NewReader(jsonData))
	resp, _ := client.Do(requetPost)
	defer resp.Body.Close()
	bodyContent, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(bodyContent))
}

func TestPostContent2(t *testing.T) {
	data := make(map[string]interface{})
	data["test"] = "TestPostContent2"
	jsonData, _ := json.Marshal(data)
	resp, _ := http.Post( portRequest+"/index", "application/json", bytes.NewReader(jsonData))
	defer resp.Body.Close()
}

func TestPostTimeout(t *testing.T) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	data := make(map[string]interface{})
	data["test"] = "TestPostTimeout"
	jsonData, _ := json.Marshal(data)
	requetPost, _ := http.NewRequest("POST", portRequest+"/index", bytes.NewReader(jsonData))
	resp, err := client.Do(requetPost)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	bodyContent, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(bodyContent))
}

func TestPostTcpTimeout(t *testing.T) {
	// 这个需要研究下
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext:(&net.Dialer{
				Timeout: 2 * time.Second,
				Deadline: time.Now().Add(3 * time.Second),
				KeepAlive: 2 * time.Second,
			}).DialContext,
		},
		Timeout: 8 * time.Second,
	}

	data := make(map[string]interface{})
	data["test"] = "TestPostTimeout"
	jsonData, _ := json.Marshal(data)
	requetPost, _ := http.NewRequest("POST", portRequest+"/index", bytes.NewReader(jsonData))
	resp, err := client.Do(requetPost)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	bodyContent, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(bodyContent))
}