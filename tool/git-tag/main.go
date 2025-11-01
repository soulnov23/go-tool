package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// v${MAJOR}.${MINOR}.${PATCH}
const (
	tagPrefix           = "v"
	initialVersion      = "1.1.0"
	minorCarryThreshold = 99
	patchCarryThreshold = 9
)

var workdir string

func main() {
	log.SetFlags(0)
	log.SetPrefix("\033[1;32m[git-tag]\033[m ")

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ 获取当前工作目录失败: %v", err)
	}
	workdir = cwd

	// 1. 获取当前版本
	currentVersion := getCurrentVersion()

	// 2. 递增版本号
	newVersion := incrementVersion(currentVersion)

	// 3. 创建Git标签
	createGitTag(newVersion)

	// 4. 推送标签到远程
	pushGitTag(newVersion)
}

// 获取当前版本
func getCurrentVersion() string {
	// 执行git命令获取最新标签
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = workdir
	if err := cmd.Run(); err != nil {
		// 没有标签时使用初始版本
		log.Printf("⚠️ 获取Git标签失败: %s", stderr.String())
		log.Printf("⚠️ 使用初始版本: %s", initialVersion)
		return initialVersion
	}

	tag := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(tag, tagPrefix) {
		log.Fatalf("❌ 无效的标签格式: %s", tag)
	}

	// 去掉标签前缀 "v"
	currentVersion := strings.TrimPrefix(tag, tagPrefix)
	log.Printf("✅ 从Git标签获取当前版本: %s", currentVersion)
	return currentVersion
}

// 递增版本号
func incrementVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		log.Fatalf("❌ 无效的版本格式: %s", version)
	}

	// 转换为整数
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Fatalf("❌ 无效的版本格式: %s", version)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf("❌ 无效的版本格式: %s", version)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Fatalf("❌ 无效的版本格式: %s", version)
	}

	// 智能进位逻辑
	patch++
	if patch > patchCarryThreshold {
		patch = 0
		minor++
		if minor > minorCarryThreshold {
			minor = 0
			major++
		}
	}
	newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	log.Printf("✅ 版本递增成功: %s", newVersion)
	return newVersion
}

// 创建Git标签
func createGitTag(version string) {
	tag := tagPrefix + version
	cmd := exec.Command("git", "tag", tag)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = workdir
	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ 创建Git标签[%s]失败: %s", tag, stderr.String())
	}
	log.Printf("✅ 已创建Git标签: %s", tag)
}

// 推送Git标签
func pushGitTag(version string) {
	tag := tagPrefix + version
	cmd := exec.Command("git", "push", "origin", tag)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = workdir
	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ 推送Git标签[%s]失败: %s", tag, stderr.String())
	}
	log.Printf("✅ 已推送Git标签: %s", tag)
}
