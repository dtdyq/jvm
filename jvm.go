package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/dtdyq/jvm/local"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

//=============define begin=================//

type CmdInfo struct {
	cmd  string
	desc string
	proc func(subs []string)
}

var commands = []CmdInfo{
	{
		cmd:  "help",
		desc: "for help info",
	},
	{
		cmd:  "on",
		desc: "enable java version manager[will remove old jdk envs]",
		proc: enableJvm,
	},
	{
		cmd:  "off",
		desc: "disable java version manager,revert old env[try]",
		proc: disableJvm,
	},
	{
		cmd:  "cur",
		desc: "current activated jdk",
		proc: currentActiveJdk,
	},
	{
		cmd:  "list",
		desc: "all installed jdk",
		proc: listInstalledJdk,
	},
	{
		cmd:  "inst",
		desc: "<version> [param] version like 21.0.1 or lts for latest lts version,\nparam:jdk vendor [openjdk|graal|oraclejdk|liberica(default)]",
		proc: instJdk,
	},
	{
		cmd:  "use",
		desc: "<version> <vendor> use the specify jdk",
		proc: useJdk,
	},
}

// vendor_version_system_arch
//
//vendor:liberica openjdk oracle graal
//version:8 11 17 21
//system:windows linux macos
//arch:x32 x64 arch64 arch32
var supportVendor = []string{"liberica", "openjdk", "oracle", "graal"}
var supportVersion = []string{"8", "11", "17", "21"}
var supportSys = []string{"windows", "linux", "macos"}
var supportArch = []string{"x32", "x64", "arch64", "arch32"}
var jdks = map[string]string{
	"liberica_21_windows_x64":    "https://download.bell-sw.com/java/21.0.2+14/bellsoft-jdk21.0.2+14-windows-amd64.zip",
	"liberica_21_windows_arch64": "https://download.bell-sw.com/java/21.0.2+14/bellsoft-jdk21.0.2+14-windows-aarch64.zip",
	"liberica_21_linux_x64":      "https://download.bell-sw.com/java/21.0.2+14/bellsoft-jdk21.0.2+14-linux-amd64.tar.gz",
	"liberica_21_linux_arch64":   "https://download.bell-sw.com/java/21.0.2+14/bellsoft-jdk21.0.2+14-linux-aarch64.tar.gz",
	"liberica_21_macos_x64":      "https://download.bell-sw.com/java/21.0.2+14/bellsoft-jdk21.0.2+14-macos-amd64.tar.gz",
	"liberica_21_macos_arch64":   "https://download.bell-sw.com/java/21.0.2+14/bellsoft-jdk21.0.2+14-macos-aarch64.tar.gz",

	"liberica_17_windows_x64":    "https://download.bell-sw.com/java/17.0.10+13/bellsoft-jdk17.0.10+13-windows-amd64.zip",
	"liberica_17_windows_arch64": "https://download.bell-sw.com/java/17.0.10+13/bellsoft-jdk17.0.10+13-windows-aarch64.zip",
	"liberica_17_linux_x64":      "https://download.bell-sw.com/java/17.0.10+13/bellsoft-jdk17.0.10+13-linux-amd64.tar.gz",
	"liberica_17_linux_arch64":   "https://download.bell-sw.com/java/17.0.10+13/bellsoft-jdk17.0.10+13-linux-aarch64.tar.gz",
	"liberica_17_macos_x64":      "https://download.bell-sw.com/java/17.0.10+13/bellsoft-jdk17.0.10+13-macos-amd64.tar.gz",
	"liberica_17_macos_arch64":   "https://download.bell-sw.com/java/17.0.10+13/bellsoft-jdk17.0.10+13-macos-aarch64.tar.gz",

	"liberica_11_windows_x64":    "https://download.bell-sw.com/java/11.0.22+12/bellsoft-jdk11.0.22+12-windows-amd64.zip",
	"liberica_11_windows_arch64": "https://download.bell-sw.com/java/11.0.22+12/bellsoft-jdk11.0.22+12-windows-aarch64.zip",
	"liberica_11_linux_x64":      "https://download.bell-sw.com/java/11.0.22+12/bellsoft-jdk11.0.22+12-linux-amd64.tar.gz",
	"liberica_11_linux_arch64":   "https://download.bell-sw.com/java/11.0.22+12/bellsoft-jdk11.0.22+12-linux-aarch64.tar.gz",
	"liberica_11_macos_x64":      "https://download.bell-sw.com/java/11.0.22+12/bellsoft-jdk11.0.22+12-macos-amd64.tar.gz",
	"liberica_11_macos_arch64":   "https://download.bell-sw.com/java/11.0.22+12/bellsoft-jdk11.0.22+12-macos-aarch64.tar.gz",

	"liberica_8_windows_x64":    "https://download.bell-sw.com/java/8u402+7/bellsoft-jdk8u402+7-windows-amd64.zip",
	"liberica_8_windows_arch64": "https://download.bell-sw.com/java/8u402+7/bellsoft-jdk8u402+7-windows-aarch64.zip",
	"liberica_8_linux_x64":      "https://download.bell-sw.com/java/8u402+7/bellsoft-jdk8u402+7-linux-amd64.tar.gz",
	"liberica_8_linux_arch64":   "https://download.bell-sw.com/java/8u402+7/bellsoft-jdk8u402+7-linux-aarch64.tar.gz",
	"liberica_8_macos_x64":      "https://download.bell-sw.com/java/8u402+7/bellsoft-jdk8u402+7-macos-amd64.tar.gz",
	"liberica_8_macos_arch64":   "https://download.bell-sw.com/java/8u402+7/bellsoft-jdk8u402+7-macos-aarch64.tar.gz",

	"openjdk_21_windows_x64":  "https://download.java.net/java/GA/jdk21/fd2272bbf8e04c3dbaee13770090416c/35/GPL/openjdk-21_windows-x64_bin.zip",
	"openjdk_21_macos_x64":    "https://download.java.net/java/GA/jdk21/fd2272bbf8e04c3dbaee13770090416c/35/GPL/openjdk-21_macos-x64_bin.tar.gz",
	"openjdk_21_macos_arch64": "https://download.java.net/java/GA/jdk21/fd2272bbf8e04c3dbaee13770090416c/35/GPL/openjdk-21_macos-aarch64_bin.tar.gz",
	"openjdk_21_linux_x64":    "https://download.java.net/java/GA/jdk21/fd2272bbf8e04c3dbaee13770090416c/35/GPL/openjdk-21_linux-x64_bin.tar.gz",
	"openjdk_21_linux_arch64": "https://download.java.net/java/GA/jdk21/fd2272bbf8e04c3dbaee13770090416c/35/GPL/openjdk-21_linux-aarch64_bin.tar.gz",

	"openjdk_17_windows_x64":  "https://download.java.net/java/GA/jdk17/0d483333a00540d886896bac774ff48b/35/GPL/openjdk-17_windows-x64_bin.zip",
	"openjdk_17_macos_x64":    "https://download.java.net/java/GA/jdk17/0d483333a00540d886896bac774ff48b/35/GPL/openjdk-17_macos-x64_bin.tar.gz",
	"openjdk_17_macos_arch64": "https://download.java.net/java/GA/jdk17/0d483333a00540d886896bac774ff48b/35/GPL/openjdk-17_macos-aarch64_bin.tar.gz",
	"openjdk_17_linux_x64":    "https://download.java.net/java/GA/jdk17/0d483333a00540d886896bac774ff48b/35/GPL/openjdk-17_linux-x64_bin.tar.gz",
	"openjdk_17_linux_arch64": "https://download.java.net/java/GA/jdk17/0d483333a00540d886896bac774ff48b/35/GPL/openjdk-17_linux-aarch64_bin.tar.gz",

	"openjdk_11_windows_x64": "https://download.java.net/java/ga/jdk11/openjdk-11_windows-x64_bin.zip",
	"openjdk_11_macos_x64":   "https://download.java.net/java/ga/jdk11/openjdk-11_osx-x64_bin.tar.gz",
	"openjdk_11_linux_x64":   "https://download.java.net/java/ga/jdk11/openjdk-11_linux-x64_bin.tar.gz",
}

const ckEnabled = "enabled"
const ckActivated = "active"

var config = map[string]string{}
var workPath = ""
var jdkPath = ""

// =============define end=================//
func init() {
	tryInitEnv()
}

func main() {
	defer exit()
	args := os.Args
	if len(args) <= 1 {
		help(commands)
		return
	}
	mc := args[1]

	if mc != "on" && mc != "off" {
		if !getBoolConfig(ckEnabled, false) {
			color.Yellow("before start,use jvm on to enable java version manager")
			return
		}
	}
	var b = false
	for _, cmd := range commands {
		if mc == "help" {
			b = true
			help(commands)
			return
		}
		if mc == cmd.cmd {
			b = true
			cmd.proc(args[2:])
			break
		}
	}
	if !b {
		help(commands)
	}
}

//================setup start===================//

func tryInitEnv() {
	dir, err := os.UserHomeDir()
	if err != nil {
		color.Red("%s", err)
		return
	}
	if !pathExist(filepath.Join(dir, ".jvm")) {
		err = os.Mkdir(filepath.Join(dir, ".jvm"), os.ModePerm)
		if err != nil {
			color.Red("create work dir err:%s", err)
			return
		}
	}
	workPath = filepath.Join(dir, ".jvm")
	jdkPath = filepath.Join(dir, ".jvm", "jdks")
	if !pathExist(jdkPath) {
		err = os.Mkdir(jdkPath, os.ModePerm)
		if err != nil {
			color.Red("create jdk dir err:%s", err)
			return
		}
	}
	cfgFile := filepath.Join(dir, ".jvm", "jvm.cfg")
	file, err := os.OpenFile(cfgFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		color.Red("%s", err)
		return
	}
	r := bufio.NewScanner(file)
	if err != nil {
		color.Red("%s", err)
		return
	}
	for r.Scan() {
		text := r.Text()
		if strings.TrimSpace(text) == "" {
			continue
		}
		kv := strings.SplitN(strings.TrimSpace(text), "=", 2)
		config[kv[0]] = kv[1]
	}
}

func exit() {
	dir, _ := os.UserHomeDir()
	cfgFile := filepath.Join(dir, ".jvm", "jvm.cfg")
	file, err := os.OpenFile(cfgFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		color.Red("exit:%s", err)
		return
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	defer w.Flush()
	for k, v := range config {
		_, err = w.WriteString(fmt.Sprintf("%s=%s\n", k, v))
		if err != nil {
			color.HiYellow("%s", err)
		}
	}
}

//=================setup end======================//

//====================util start==========================//

func pathExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func getConfig(key string, def string) string {
	val, exist := config[key]
	if !exist {
		return def
	}
	return val
}

func getBoolConfig(key string, def bool) bool {
	ret, err := strconv.ParseBool(getConfig(key, strconv.FormatBool(def)))
	if err != nil {
		return def
	}
	return ret
}

func getI32Config(key string, def int32) int32 {
	ret, err := strconv.Atoi(getConfig(key, strconv.Itoa(int(def))))
	if err != nil {
		return def
	}
	return int32(ret)
}

func downloadKeyBy(version string, vendor string) string {
	system := runtime.GOOS
	if system == "darwin" {
		system = "macos"
	}
	arch := runtime.GOARCH
	switch arch {
	case "386":
		arch = "x32"
		break
	case "amd64":
		arch = "x64"
		break
	case "arm":
		arch = "arch32"
		break
	case "arm64":
		arch = "arch64"
		break
	}
	return fmt.Sprintf("%s_%s_%s_%s", vendor, version, system, arch)
}

func extractRelevantDirs(tarball string) error {
	file, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer file.Close()
	target := filepath.Dir(tarball)
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		color.Red("%s", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	var relevantDirPrefix string

	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			return nil // End of archive
		case err != nil:
			color.Red("%s", err)
			return err
		case header == nil:
			continue
		}
		name := header.Name
		if relevantDirPrefix == "" && strings.Contains(name, "/bin/") {
			relevantDirPrefix = strings.Split(name, "/bin/")[0]
			continue
		}
		if relevantDirPrefix != "" && strings.HasPrefix(name, relevantDirPrefix) {
			targetPath := filepath.Join(target, strings.TrimPrefix(name, relevantDirPrefix))

			switch header.Typeflag {
			case tar.TypeDir:
				if err := os.MkdirAll(targetPath, 0755); err != nil {
					return err
				}
			case tar.TypeReg:
				if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
					color.Red("%s", err)
					return err
				}
				outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
				if err != nil {
					color.Red("%s", err)
					return err
				}
				if _, err := io.Copy(outFile, tarReader); err != nil {
					outFile.Close()
					color.Red("%s", err)
					return err
				}
				outFile.Close()
			}
		}
	}
}

// unzipJDK 提取ZIP文件中的 'bin' 目录和它同级的其他目录或文件。
func unzipJDK(src string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	dest := filepath.Dir(src)
	// 找到 'bin' 目录在ZIP文件中的路径
	var binPath string
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "/bin/") {
			binPath = strings.TrimSuffix(f.Name, "bin/")
			break
		}
	}

	if binPath == "" {
		return fmt.Errorf("没有在ZIP文件中找到 'bin' 目录")
	}

	// 解压 'bin' 目录同级的所有文件和目录
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, binPath) {
			fPath := filepath.Join(dest, strings.TrimPrefix(f.Name, binPath))
			if f.FileInfo().IsDir() {
				os.MkdirAll(fPath, os.ModePerm)
				continue
			}

			if err = os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
				return err
			}

			outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			rc, err := f.Open()
			if err != nil {
				return err
			}

			_, err = io.Copy(outFile, rc)

			outFile.Close()
			rc.Close()

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func downloadJdkTo(url, key string) {
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	var tail = ".zip"
	if strings.HasSuffix(url, "gz") {
		tail = ".tar.gz"
	}

	sp := filepath.Join(jdkPath, key)
	if !pathExist(sp) {
		err := os.Mkdir(sp, os.ModePerm)
		if err != nil {
			color.Red("create work dir err:%s", err)
			return
		}
	}

	f, err := os.OpenFile(filepath.Join(sp, key+tail), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		color.Red("do write error:%s", err)
	}
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading ",
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)
	if strings.Contains(key, "windows") {
		color.White("download done.start unzip...")
		unzipJDK(filepath.Join(sp, key+tail))
	} else {
		color.White("download done.start extract...")
		extractRelevantDirs(filepath.Join(sp, key+tail))
	}
	color.Green("install jdk success:%s", key)
}

func changeEnvSymbol(key string) {
	originalPath := filepath.Join(jdkPath, key)
	var symlinkPath = local.JdkHomeLinkPath
	if _, err := os.Lstat(symlinkPath); err == nil {
		if err = os.Remove(symlinkPath); err != nil {
			color.Red("remove old version link fail:%s", err)
			return
		}
	}
	err := os.Symlink(originalPath, symlinkPath)
	if err != nil {
		color.Red("active new version fail:%s", err)
		return
	} else {
		if runtime.GOOS == "windows" {
			exePath := filepath.Join(jdkPath, key, "bin")
			var exeSymPath = local.JdkExeLinkPath
			if _, err := os.Lstat(exeSymPath); err == nil {
				if err = os.Remove(exeSymPath); err != nil {
					color.Red("remove old version link fail:%s", err)
					return
				}
			}
			err := os.Symlink(exePath, exeSymPath)
			if err != nil {
				color.Red("active new version fail:%s", err)
			}
		}
	}

	p := strings.Split(key, "_")
	color.Green("active %s %s success", p[0], p[1])
	color.Green("use java --version find out")
	color.Green("use jvm list show all installed jdks")
	config[ckActivated] = key
}

//====================util end==========================//

//====================cmd start==========================//

func help(commands []CmdInfo) {
	color.White("Usage:")
	color.White("")
	var cmdLen = -1
	for _, cmd := range commands {
		l := len(fmt.Sprintf("  jvm %s  ", cmd.cmd))
		if l > cmdLen {
			cmdLen = l
		}
	}
	for _, cmd := range commands {
		c := fmt.Sprintf("  jvm %s", cmd.cmd)
		c = c + strings.Repeat(" ", cmdLen-len(c))
		color.New(color.FgCyan).Print(c)
		for idx, line := range strings.Split(cmd.desc, "\n") {
			if idx == 0 {
				color.White(" : %s", line)
			} else {
				color.White(strings.Repeat(" ", cmdLen) + fmt.Sprintf("   %s", line))
			}
		}
	}
}
func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func useJdk(subs []string) {
	if subs == nil || len(subs) == 0 {
		color.Yellow("version required,use [jvm help] for detail")
		return
	}
	version := subs[0]
	if !contains(supportVersion, version) {
		color.Red("un support version:%s, use [jvm detail] for help", version)
		return
	}
	var vendor = "liberica"
	if len(subs) > 1 {
		if !contains(supportVendor, subs[1]) {
			color.Red("un support jdk type:%s, use [jvm detail] for help", subs[1])
			return
		} else {
			vendor = subs[1]
		}
	}
	key := downloadKeyBy(version, vendor)
	if !pathExist(filepath.Join(jdkPath, key)) {
		color.Red("current version not install try jvm inst <version> [param] first")
		return
	}
	changeEnvSymbol(key)
}

func instJdk(subs []string) {
	if subs == nil || len(subs) == 0 {
		color.Yellow("missing param:<version>")
		return
	}
	version := subs[0]
	if !contains(supportVersion, version) {
		color.Red("un support version:%s, use [jvm detail] for help", version)
		return
	}
	var vendor = "liberica"
	if len(subs) > 1 {
		if !contains(supportVendor, subs[1]) {
			color.Red("un support jdk type:%s, use [jvm detail] for help", subs[1])
			return
		} else {
			vendor = subs[1]
		}
	}
	key := downloadKeyBy(version, vendor)
	fmt.Println(key)
	url, exist := jdks[key]
	if !exist {
		color.Red("not support for %s,use [jvm detail] for help", key)
		return
	}
	downloadJdkTo(url, key)
}

func currentActiveJdk(subs []string) {
	act := getConfig(ckActivated, "")
	if act == "" {
		color.Yellow("no jdk activated;use [jvm inst] to install,use [jvm use] to active")
	} else {
		ns := strings.Split(act, "_")
		color.Magenta("  %s [%s]", ns[1], ns[0])
	}
}

func listInstalledJdk(subs []string) {
	entries, err := os.ReadDir(jdkPath)
	if err != nil {
		color.Red("list jdks failed:%s", err)
		return
	}
	act := getConfig(ckActivated, "")
	for _, e := range entries {
		if e.IsDir() && strings.Count(e.Name(), "_") == 3 {
			name := strings.TrimSpace(e.Name())
			ns := strings.Split(name, "_")
			if act == name {
				color.Magenta("  %s [%s] current active", ns[1], ns[0])
			} else {
				color.Blue("  %s [%s] ", ns[1], ns[0])
			}
		}
	}
}

func disableJvm(subs []string) {
	config[ckEnabled] = "false"
	color.Green("jvm enabled try [jvm inst <version> or jvm use <version>] to use")

}

func enableJvm(subs []string) {
	local.SetupJavaHomeAndPath()
	config[ckEnabled] = "true"
	color.Green("jvm enabled try [jvm inst <version> or jvm use <version>] to use")
}

//======================cmd end=======================//
