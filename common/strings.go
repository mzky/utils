package common

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Len 字符串长度，汉字算1个
func Len(str string) int {
	return utf8.RuneCountInString(str)
}

// Or 当input为空值时返回默认值default
func Or(input, defaultVal string) string {
	if input == "" {
		return defaultVal
	}
	return input
}

// OrValue 当input为指定值时返回默认值defaultVal
func OrValue(input, seq, defaultVal string) string {
	if input == seq {
		return defaultVal
	}
	return input
}

// NotValue 当input不等于指定值时返回默认值defaultVal
func NotValue(input, seq, defaultVal string) string {
	if input != seq {
		return defaultVal
	}
	return input
}

// StringPrefixEqualCount 判断连续相同的字符串个数
func StringPrefixEqualCount(str1, str2 string) int {
	len1 := len(str1)
	len2 := len(str2)

	var length int
	if len1 > len2 {
		length = len2
	} else {
		length = len1
	}
	count := 0
	for i := 0; i < length; i++ {
		if str1[i] == str2[i] {
			count++
		}
	}
	return count
}

// ReplaceFileContent 使用正则表达式查找模式，并且替换正则1号捕获分组为指定的内容
func ReplaceFileContent(filename, regexStr, repl string) error {
	conf, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("ReadFile %s error %w", filename, err)
	}

	fixed, err := ReplaceRegexGroup1(string(conf), regexStr, repl)
	if err != nil {
		return err
	}

	stat, _ := os.Stat(filename)

	return os.WriteFile(filename, []byte(fixed), stat.Mode())
}

// ReplaceFileKeywords 字符串替换，并且替换正则1号捕获分组为指定的内容
func ReplaceFileKeywords(filename, str, repl string) error {
	conf, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("ReadFile %s error %w", filename, err)
	}

	reg := regexp.MustCompile(str)
	newContent := reg.ReplaceAllString(string(conf), repl)

	stat, _ := os.Stat(filename)

	return os.WriteFile(filename, []byte(newContent), stat.Mode())
}

// SearchFileContent 使用正则表达式查找模式正则1号捕获分组
func SearchFileContent(filename, regexStr string) ([]string, error) {
	conf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ReadFile %s error %w", filename, err)
	}

	return FindRegexGroup1(string(conf), regexStr)
}

// SearchPatternLinesInFile 使用正则表达式boundaryRegexStr在文件filename中查找大块，
// 然后在大块中用captureGroup1Regex中的每行寻找匹配
func SearchPatternLinesInFile(filename, boundaryRegexStr, captureGroup1Regex string) ([]string, error) {
	str, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ReadFile %s error %w", filename, err)
	}

	return SearchPatternLines(string(str), boundaryRegexStr, captureGroup1Regex)
}

// SearchPatternLines 使用正则表达式boundaryRegexStr在str中查找大块，
// 然后在大块中用captureGroup1Regex中的每行寻找匹配
func SearchPatternLines(str, boundaryRegexStr, captureGroup1Regex string) ([]string, error) {
	founds, err := FindRegexGroup1(str, boundaryRegexStr)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0)

	for _, v := range founds {
		vv, err := FindRegexGroup1(v, captureGroup1Regex)
		if err != nil {
			return nil, err
		}

		lines = append(lines, vv...)
	}

	return lines, nil
}

// FindRegexGroup1 使用正则表达式regexStr在str中查找内容
func FindRegexGroup1(str, regexStr string) ([]string, error) {
	re, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, err
	}

	group1s := make([]string, 0)

	for _, v := range re.FindAllStringSubmatch(str, -1) {
		if len(v) < 2 { // nolint gomnd
			return nil, fmt.Errorf("regexp %s should have at least one capturing group", regexStr)
		}

		group1s = append(group1s, v[1])
	}

	return group1s, nil
}

// ReplaceRegexGroup1 使用正则表达式regexStr在str中查找内容，并且替换正则1号捕获分组为指定的内容
func ReplaceRegexGroup1(str, regexStr, repl string) (string, error) {
	re, err := regexp.Compile(regexStr)
	if err != nil {
		return "", err
	}

	fixed := ""
	lastIndex := 0

	for _, v := range re.FindAllStringSubmatchIndex(str, -1) {
		if len(v) < 4 { // nolint gomnd
			return "", fmt.Errorf("regexp %s should have at least one capturing group", regexStr)
		}

		fixed += str[lastIndex:v[2]] + repl
		lastIndex = v[3]
	}

	if lastIndex == 0 {
		return "", fmt.Errorf("regexp %s found non submatches", regexStr)
	}

	return fixed + str[lastIndex:], nil
}

// StripUnprintable 仅使用可见字符
func StripUnprintable(str string) string {
	var b strings.Builder
	b.Grow(len(str))

	for _, ch := range str {
		if unicode.IsPrint(ch) && !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}
