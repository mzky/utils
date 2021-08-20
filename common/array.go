package common

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// GetNumberValueArray 从字符串数组中获取数字，取数组[0]，没有数字会报错
func GetNumberValueArray(getValArray []string) ([]int, error) {
	retIntArray := make([]int, 0)
	reg := regexp.MustCompile(`[0-9]+`)

	for _, strVal := range getValArray {
		numArray := reg.FindAllString(strVal, -1) //-1表示全部匹配
		if len(numArray) < 1 {
			return nil, errors.New("正则表达式未找到数字")
		}
		num, err := strconv.Atoi(numArray[0])
		if err != nil {
			return nil, errors.New("字符串转数字失败")
		}
		retIntArray = append(retIntArray, num)
	}
	return retIntArray, nil
}

// TrimValueFromArray 去除数组中指定元素
func TrimValueFromArray(strArray []string, trimValue string) []string {
	newArray := make([]string, 0)
	for _, v := range strArray {
		if strings.TrimSpace(trimValue) != strings.TrimSpace(v) {
			newArray = append(newArray, strings.TrimSpace(v))
		}
	}

	return newArray
}

// ContainsDuplicate 检测数组是否包含重复元素 重复返回true
func ContainsDuplicate(str []string) bool {
	hash := make(map[string]bool)
	for _, v := range str {
		if hash[v] == true {
			return true
		} else {
			hash[v] = true
		}
	}
	return false
}

// RemoveRepeatedElement 删除重复元素
func RemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	sort.Strings(arr)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

// ArrayContains array中是否存在指定数据
func ArrayContains(array []string, val string) bool {
	for _, v := range array {
		if strings.TrimSpace(v) == strings.TrimSpace(val) {
			return true
		}
	}
	return false
}

// ArrayContainsNums 统计连个数组中相同值的个数
func ArrayContainsNums(s []string, substr []string) int {
	num := 0
	for _, sub := range substr {
		for _, v := range s {
			if sub == v {
				num++
			}
		}
	}
	return num
}
