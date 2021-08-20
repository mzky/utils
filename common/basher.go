package common

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

func ShellExec(shellPath string) (string, error) {
	//Start 非阻塞
	//Run   阻塞
	//替换脚本中的^M，避免异常情况
	if !FileIsExist(shellPath) {
		logrus.Errorf("文件不存在")
		return "", errors.New("文件不存在")
	}
	_ = exec.Command("/usr/bin/sed", "-i", `'s/\r//g'`, shellPath).Start()

	//附执行权限，避免异常情况
	_ = os.Chmod(shellPath, 0755)

	//脚本exit -1时run中断报错
	return CommandContext("/bin/bash", "-c", shellPath)

}

// CommandContext 执行命令 默认超时时间60秒
func CommandContext(name string, arg ...string) (string, error) {
	//设置超时时间
	ctxt, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.CommandContext(ctxt, name, arg...)

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	//Run执行c包含的命令，并阻塞直到完成. 这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	if err := cmd.Start(); err != nil {
		logrus.Errorf("%+v", err)
		return out.String(), err
	}

	if err := cmd.Wait(); err != nil {
		logrus.Errorf("%+v", err)
		return out.String(), err
	}

	logrus.Debugf("%+v\n", out.String())
	return out.String(), nil
}
