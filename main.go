package main

import (
	// "net/http"
	// // "strconv"
	"fmt"
	"os"
	// // "regexp"
	// // "strings"
	// "math/rand"
	// "time"
	"sync"

	// "github.com/mozillazg/request"
	"bytes"
	"encoding/json"
	"runtime"
	"unicode"

	"github.com/panjf2000/ants"
)

type NLPData struct {
	Id        uint32 `json:"id"`
	UniqueKey string `json:"uniqueKey"`
	TitleUkey string `json:"titleUkey"`
	DataType  string `json:"dataType"`
	Title     string `json:"title"`
	Content   string `json:"content"`
}

var punctuationMap = map[rune]rune{
	'.': '。',
	',': '，',
	':': '：',
	';': '；',
	'!': '！',
	'?': '？',
	'(': '（',
	')': '）',
}

func convertPunctuation(s *string) string {
	var result bytes.Buffer

	var lastC rune
	for i, c := range *s {
		if newC, ok := punctuationMap[c]; ok {
			if i > 0 && unicode.Is(unicode.Han, lastC) {
				c = newC
			}
		}
		result.WriteRune(c)
		lastC = c
	}
	return result.String()
}

// 判断文件夹是否存在
func HasDir(path string) (bool, error) {
	_, _err := os.Stat(path)
	if _err == nil {
		return true, nil
	}
	if os.IsNotExist(_err) {
		return false, nil
	}
	return false, _err
}

// 创建文件夹
func CreateDir(path string) {
	_exist, _err := HasDir(path)
	if _err != nil {
		fmt.Printf("获取文件夹异常 -> %v\n", _err)
		return
	}
	if _exist {
		fmt.Println("文件夹已存在！")
	} else {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			fmt.Printf("创建目录异常 -> %v\n", err)
		} else {
			fmt.Println("创建成功!")
		}
	}
}

func convert(path string, name string, wg *sync.WaitGroup) func() {
	return func() {
		file, _ := os.ReadFile(path + name)
		println(file)
		var data []NLPData
		json.Unmarshal(file, &data)
		for i, v := range data {
			data[i].Title = convertPunctuation(&v.Title)
			data[i].Content = convertPunctuation(&v.Content)
		}
		filered, _ := json.MarshalIndent(data, "", "  ")
		os.WriteFile("filtered/"+name, filered, 0666)
		wg.Done()
	}
}

func main() {

	if len(os.Args) < 2 {
		println("please input data path.")
		return
	}

	path := os.Args[1]
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	fileInfoList, err := os.ReadDir(path)
	if err != nil {
		println(err.Error())
		return
	}

	wg := sync.WaitGroup{}
	//申请一个协程池对象
	pool, _ := ants.NewPool(runtime.NumCPU())
	//关闭协程池
	defer pool.Release()
	CreateDir("filtered/")

	for _, info := range fileInfoList {
		wg.Add(1)
		pool.Submit(convert(path, info.Name(), &wg))
	}
	wg.Wait()

}
