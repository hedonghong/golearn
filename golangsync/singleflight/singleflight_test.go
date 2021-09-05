package singleflight

import (
	"fmt"
	"github.com/pingcap/errors"
	"golang.org/x/sync/singleflight"
	"sync"
	"testing"
)

var errorNotExist = errors.New("not exist")

func TestMockExample(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)

	//模拟10个并发
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			//data, err := getData("key")
			data, err := getDataUseSingleFlight("key")
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(data)
		}()
	}
	wg.Wait()
}

//获取数据
func getData(key string) (string, error) {
	data, err := getDataFromCache(key)
	if err == errorNotExist {
		//模拟从db中获取数据
		data, err = getDataFromDB(key)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		//TOOD: set cache
	} else if err != nil {
		return "", err
	}
	return data, nil
}

//模拟从cache中获取值，cache中无该值
func getDataFromCache(key string) (string, error) {
	return "", errorNotExist
}
//模拟从数据库中获取值
func getDataFromDB(key string) (string, error) {
	fmt.Printf("get %s from database\n", key)
	return "data", nil
}

var gsf singleflight.Group

func getDataUseSingleFlight(key string) (string, error) {
	data, err := getDataFromCache(key)
	if err == errorNotExist {
		//模拟从db中获取数据
		v, err, _ := gsf.Do(key, func() (interface{}, error) {
			return getDataFromDB(key)
			//set cache
		})
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		//TOOD: set cache
		data = v.(string)
	} else if err != nil {
		return "", err
	}
	return data, nil
}
