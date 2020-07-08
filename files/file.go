package util

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//获取当前目录
func SelfPath() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

//获取指定目录下的所有文件和目录,不包含子目录
func FilesAndDirs(fp, filter string) (files []string, dirs []string, err error) {
	dir, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, nil, err
	}

	PthSep := string(os.PathSeparator)
	filter = strings.ToUpper(filter) //忽略后缀匹配的大小写

	for _, fi := range dir {
		dp := path.Join(fp, PthSep, fi.Name())
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dp)
			FilesAndDirs(dp, filter)
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
func FindAllFiles(fp, filter string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	filter = strings.ToUpper(filter) //忽略后缀匹配的大小写

	for _, fi := range dir {
		dp := path.Join(fp, PthSep, fi.Name())
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dp)
			FindAllFiles(dp, filter)
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
		temp, _ := FindAllFiles(table, filter)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return files, nil
}

//判断文件夹或文件是否存在
func IsExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

//获取文件修改时间 返回unix时间戳
func FileModTime(fp string) (int64, error) {
	f, err := os.Stat(fp)
	if err != nil {
		return 0, err
	}
	return f.ModTime().Unix(), nil
}

//return 目录,文件名,后缀
func FilePathInfo(fp string) (string, string, string) {
	//filepath.Base(path)
	dir, name := filepath.Split(fp)    //目录、文件名
	return dir, name, filepath.Ext(fp) //后缀
}

func FileSize(fp string) (int64, error) {
	f, e := os.Stat(fp)
	if e != nil {
		return 0, e
	}
	return f.Size(), nil
}
