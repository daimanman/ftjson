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
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	K = flag.String("K", "1,2", "K=1")
	V = flag.Int("V", 0, "V=0")
	F = flag.String("F", " ", "空格分割符号")
	P = flag.Bool("P", false, "格式化")
	h = flag.Bool("h", false, "帮助信息")
)

var useage = `
    -K key 列
    -V value 列
    -F 列分割符号 默认 空格
    -P pretty json 输出 默认false

`

var maxCol int
var jsonMap map[string]string
var wg sync.WaitGroup
var kcols []int

func max(k []int, v int) int {
	sort.Ints(k)
	size := len(k)
	index := size - 1
	if index > 0 && k[index] > v {
		return k[index]
	}
	return v
}

func initCols() {
	kcols = make([]int, 0)
	cols := strings.Split(*K, ",")
	for _, c := range cols {
		i, _ := strconv.Atoi(c)
		kcols = append(kcols, i)
	}
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
		if size < *V {
			continue
		}

		value := strings.TrimSpace(strsArray[*V-1])
		for _, kc := range kcols {
			if kc <= size {
				key := strings.TrimSpace(strsArray[kc-1])
				if key != "" {
					jsonMap[key] = value
				}
			}
		}

	}
}

func main() {
	flag.Parse()
	if *h {
		fmt.Println(useage)
		return
	}
	initCols()
	maxCol = max(kcols, *V)
	if maxCol <= 0 {
		fmt.Println(maxCol, "请正确指定列 K V ")
		return
	}
	files := GetFiles(flag.Args())
	jsonMap = make(map[string]string)

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
