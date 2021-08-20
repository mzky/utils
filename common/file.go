package common

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FilesAndDirs 获取指定目录下的所有文件和目录,不包含子目录
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

// FindAllFiles 获取指定目录下的所有文件,包含子目录下的文件
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

// SelfPath 获取当前目录
func SelfPath() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

// IsExist 判断文件夹或文件是否存在
func IsExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

// FileModTime 获取文件修改时间 返回unix时间戳
func FileModTime(fp string) (int64, error) {
	f, err := os.Stat(fp)
	if err != nil {
		return 0, err
	}
	return f.ModTime().Unix(), nil
}

// FilePathInfo return 目录,文件名,后缀
func FilePathInfo(fp string) (string, string, string) {
	//filepath.Base(path)
	dir, name := filepath.Split(fp)    //目录、文件名
	return dir, name, filepath.Ext(fp) //后缀
}

// FileSize 获取文件大小
func FileSize(fp string) (int64, error) {
	f, e := os.Stat(fp)
	if e != nil {
		return 0, e
	}
	return f.Size(), nil
}

// FileIsExist 判断文件夹或文件是否存在, true为存在
func FileIsExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

// PathInfo 获取目录,文件名,后缀
func PathInfo(fp string) (string, string, string) {
	// filepath.Base(path)
	dir, name := filepath.Split(fp)    //目录、文件名
	return dir, name, filepath.Ext(fp) //后缀
}

// FileListFromPath 获取文件夹下文件列表，支持通配符*
func FileListFromPath(fp string) ([]string, error) {
	var fileList []string
	var err error
	if strings.Contains(fp, "*") {
		fileList, err = filepath.Glob(fp)
	} else {
		fileList, err = filepath.Glob(filepath.Join(fp, "*"))
	}

	for _, v := range fileList {
		if !IsFile(v) {
			fileList = TrimValueFromArray(fileList, v) // 删除此目录项
			fl, _ := FileListFromPath(v)
			fileList = append(fileList, fl...)
		}
	}
	return fileList, err
}

// IsFile 判断是文件还是目录
func IsFile(fp string) bool {
	fi, e := os.Stat(fp)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

// 判断所给路径文件/文件夹是否存在(返回true是存在)
func isExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// CreateMutiDir 调用os.MkdirAll递归创建文件夹
func CreateMutiDir(filePath string) error {
	if !isExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

// CleanFile 清空文件
func CleanFile(path string) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString("")
	_ = f.Close()
}

// PathFileExists 判断文件路径是否存在
func PathFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// FileExists 检查文件是否存在，并且不是目录
func FileExists(filename string) error {
	if fi, err := os.Stat(filename); err != nil {
		return err
	} else if fi.IsDir() {
		return fmt.Errorf("file %s is a directory", filename)
	}

	return nil
}

func WriteFile(filePathName, content string) error {
	file, err := os.OpenFile(filePathName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return errors.New("打开文件失败")
	}
	defer func() { _ = file.Close() }()

	buf := bufio.NewWriter(file)
	_, _ = buf.WriteString(content)
	_ = buf.Flush()

	return nil
}

func ReadFile(filePathName string) (string, error) {
	b, err := ioutil.ReadFile(filePathName)
	if err != nil {
		return "", errors.New("读取文件失败")
	}
	return string(b), nil
}

func IsFileName(fileName string) bool {
	str := "/tmp/tmpFile-" + fileName
	fp, err := os.Create(str)
	if err != nil {
		return false
	}

	err = fp.Close()
	_ = os.Remove(str)

	return true
}
