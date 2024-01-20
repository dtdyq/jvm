//go:build windows

package local

import (
	"github.com/fatih/color"
	"golang.org/x/sys/windows/registry"
	"strings"
)

const JdkHomeLinkPath = "C:\\Program Files\\jvmdkhome"
const JdkExeLinkPath = "C:\\Program Files\\jvmdkhome\\bin"

func SetupJavaHomeAndPath() {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		color.Red("read reg err:%s", err)
		return
	}
	defer k.Close()

	// 读取旧的Path值
	oldPath, _, err := k.GetStringValue("JAVA_HOME")

	// 如果旧的Path中已经包含了新目录，则不作操作
	if strings.Contains(oldPath, JdkHomeLinkPath) {
		return
	}
	err = k.SetStringValue("JAVA_HOME", JdkHomeLinkPath)
	if err != nil {
		color.Red("read jh err:%s", err)
		return
	}

	// 读取旧的Path值
	po, _, err := k.GetStringValue("Path")
	if err != nil {
		color.Red("read path err:%s", err)
		return
	}

	// 如果旧的Path中已经包含了新目录，则不作操作
	if strings.Contains(po, "%JAVA_HOME%\\bin;") {
		return
	}
	po = JdkExeLinkPath + ";" + po
	err = k.SetStringValue("Path", po)
	if err != nil {
		color.Red("set path err:%s", err)
		return
	}

}
