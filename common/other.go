package common

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"os"
	"strings"
)

func GetUUID() string {
	str := fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))
	return str
}

// GetMaxValue 获取最大值
func GetMaxValue(numVal []int) (int, error) {
	if len(numVal) < 1 {
		return 0, errors.New("数组长度不能为0")
	}

	maxVal := numVal[0]
	for _, v := range numVal {
		if v > maxVal {
			maxVal = v
		}
	}

	return maxVal, nil
}

// GetMinValue 获取最小值
func GetMinValue(numVal []int) (int, error) {
	if len(numVal) < 1 {
		return 0, errors.New("数组长度不能为0")
	}

	minVal := numVal[0]
	for _, v := range numVal {
		if v < minVal {
			minVal = v
		}
	}

	return minVal, nil
}

// ReadLine 读取指定行的内容 0行开始
func ReadLine(lineNumber int, path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	lineCount := 0
	for fileScanner.Scan() {
		if lineCount == lineNumber {
			return fileScanner.Text()
		}
		lineCount++
	}

	return ""
}

// FindLineIndex 根据关键字查找字符串索引行 返回行号
func FindLineIndex(ctx, sep string) (int, error) {
	sc := bufio.NewScanner(bytes.NewReader([]byte(ctx)))
	line := 0
	for sc.Scan() {
		if strings.Contains(sc.Text(), sep) {
			return line, nil
		}
		line++
	}
	return 0, errors.New("找不到关键字")
}

// ConfigViper 读指定位置配置文件
func ConfigViper(path string) (*viper.Viper, error) {
	if !FileIsExist(path) {
		fmt.Errorf("文件 %s 不存在", path)
		return nil, fmt.Errorf("文件 %s 不存在", path)
	}
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("打开配置文件失败,路径:%s,失败原因:%v", path, err)
	}
	return v, nil
}
