package util

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

//获取指定目录下的所有文件和目录
func GetFilesAndDirs(dirPth, filter string) (files []string, dirs []string, err error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, nil, err
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		dp := path.Join(dirPth, PthSep, fi.Name())
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dp)
			GetFilesAndDirs(dp, filter)
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), filter)
			if ok {
				files = append(files, dp)
			}
		}
	}

	return files, dirs, nil
}

//获取指定目录下的所有文件,包含子目录下的文件
func GetAllFiles(dirPth, filter string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		dp := path.Join(dirPth, PthSep, fi.Name())
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dp)
			GetAllFiles(dp, filter)
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), filter)
			if ok {
				files = append(files, dp)
			}
		}
	}

	// 读取子目录下文件
	for _, table := range dirs {
		temp, _ := GetAllFiles(table, filter)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return files, nil
}

//废弃
func FindFile(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

//判断文件夹或文件是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsExist(err)
}

//获取文件修改时间 返回unix时间戳
func GetFileModTime(path string) time.Time {
	f, err := os.Open(path)
	if err != nil {
		log.Println("打开文件失败")
		return time.Now()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Println("无法获取文件信息")
		return time.Now()
	}

	return fi.ModTime()
}
