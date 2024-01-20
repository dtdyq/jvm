package local

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const JdkHomeLinkPath = "/usr/local/jvmjdkhome"
const JdkExeLinkPath = "/usr/local/jvmjdkhome/bin"

func SetupJavaHomeAndPath() {

	// 获取当前用户信息
	currentUser, err := user.Current()
	if err != nil {
		color.Red("get user info err:%s", err)
		return
	}

	// 确定要修改的文件
	var ep = ""

	// 检查文件是否存在
	_, err = os.Stat(filepath.Join(currentUser.HomeDir, ".bashrc"))
	if err == nil {
		ep = filepath.Join(currentUser.HomeDir, ".bashrc")
	} else {
		color.Red("error:not found env profile:%s", err)
		return

	}
	if ep == "" {
		_, err = os.Stat(filepath.Join(currentUser.HomeDir, ".bash_profile"))
		if err == nil {
			ep = filepath.Join(currentUser.HomeDir, ".bash_profile")
		} else {
			color.Red("error:not found env profile:%s", err)
			return
		}
	}
	// 检查文件中是否已存在该路径
	foundPath := checkPathInFile(ep, fmt.Sprintf("export JAVA_HOME=%s", JdkHomeLinkPath))
	if foundPath {
		return
	}

	// 要执行的命令
	cmdString := fmt.Sprintf("echo 'export JAVA_HOME=%s' >> %s\necho 'export PATH=$JAVA_HOME/bin:$PATH' >> %s\nsource %s\n", JdkHomeLinkPath, ep, ep, ep)
	cmd := exec.Command("bash", "-c", cmdString)

	// 执行命令
	err = cmd.Run()
	if err != nil {
		color.Red("setup path error:%s", err)
		return
	}
}

// checkPathInFile 检查配置文件中是否已存在pathToAdd的路径
func checkPathInFile(filePath, pathToAdd string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("打开文件时发生错误: %s\n", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// 如果找到一个包含我们想要添加路径的行，说明路径已存在
		if line := scanner.Text(); strings.Contains(line, pathToAdd) {
			return true
		}
	}
	if err := scanner.Err(); err != nil {
		return false
	}
	return false
}
