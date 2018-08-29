package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

var (
	K = flag.Int("K", 0, "K=0")
	V = flag.Int("V", 0, "V=0")
	F = flag.String("F", " ", "空格分割符号")
	P = flag.Bool("P", true, "格式化")
)

var maxCol int
var jsonMap map[string]string
var wg sync.WaitGroup

func max(k int, v int) int {
	if k > v {
		return k
	}
	return v
}

func GetFiles(paths []string) []string {
	fs := make([]string, 0)
	for _, p := range paths {
		dir := path.Dir(p)
		base := path.Base(p)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Println(dir, "文件夹不存在", err.Error())
			continue
		}

		for _, file := range files {
			fname := file.Name()
			isDir := file.IsDir()
			if isDir {
				continue
			}
			match, _ := path.Match(base, fname)
			if match {
				fs = append(fs, path.Join(dir, fname))
			}
		}
	}
	return fs
}

func dealFile(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer func() {
		file.Close()
		wg.Done()
	}()
	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}

		strsArray := strings.Split(line, *F)
		size := len(strsArray)
		if size < maxCol {
			continue
		}
		key := strings.TrimSpace(strsArray[*K-1])
		value := strings.TrimSpace(strsArray[*V-1])
		jsonMap[key] = value
	}
}

func main() {
	flag.Parse()
	files := GetFiles(flag.Args())
	jsonMap = make(map[string]string)
	maxCol = max(*K, *V)
	lenth := len(files)
	if len(files) == 0 {
		fmt.Printf("未找到文件,请检查参数是否在正确 \n")
		return
	}

	wg.Add(lenth)
	for _, fpath := range files {
		go dealFile(fpath)
	}
	wg.Wait()

	bs, err := json.Marshal(jsonMap)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if *P {
		var out bytes.Buffer
		err = json.Indent(&out, bs, "", "  ")
		out.WriteTo(os.Stdout)
	} else {
		fmt.Printf("%s\n", string(bs))
	}

}
