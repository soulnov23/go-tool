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
	initialVersion      = "1.0.0"
	minorCarryThreshold = 99
	patchCarryThreshold = 9
)

var workdir string

func main() {
	log.SetFlags(0)
	log.SetPrefix("\033[1;32m[gittag]\033[m ")

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ 获取当前工作目录失败: %v", err)
	}
	workdir = cwd

	// 1. 检查当前HEAD是否已经打过tag，如果是则跳过
	if tags := getHeadTags(); len(tags) > 0 {
		log.Fatalf("📢 当前最新提交已有标签%v，跳过打标签", tags)
	}

	// 2. 获取当前版本
	currentVersion := getCurrentVersion()

	// 3. 递增版本号
	newVersion := incrementVersion(currentVersion)

	// 4. 创建并推送Git标签
	createAndPushGitTag(newVersion)
}

// 获取当前HEAD commit上的所有tag
func getHeadTags() []string {
	cmd := exec.Command("git", "tag", "--points-at", "HEAD")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = workdir
	if err := cmd.Run(); err != nil {
		return nil
	}
	var tags []string
	for tag := range strings.SplitSeq(strings.TrimSpace(stdout.String()), "\n") {
		tag = strings.TrimSpace(tag)
		if tag != "" && strings.HasPrefix(tag, tagPrefix) {
			tags = append(tags, tag)
		}
	}
	return tags
}

// 获取当前版本
func getCurrentVersion() string {
	// 执行git命令获取最新标签
	cmd := exec.Command("git", "tag", "--sort=-version:refname")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = workdir
	if err := cmd.Run(); err == nil {
		for tag := range strings.SplitSeq(strings.TrimSpace(stdout.String()), "\n") {
			tag = strings.TrimSpace(tag)
			currentVersion, ok := strings.CutPrefix(tag, tagPrefix)
			if !ok {
				log.Printf("📢 跳过无效标签[%s]: %v", tag, err)
				continue
			}
			if _, _, _, err := parseVersion(currentVersion); err != nil {
				log.Printf("📢 跳过无效标签[%s]: %v", tag, err)
				continue
			}
			log.Printf("✅ 从Git标签获取当前版本[%s]", currentVersion)
			return currentVersion
		}
	}

	log.Printf("📢 未找到任何有效Git标签，使用初始版本[%s]", initialVersion)
	return initialVersion
}

// parseVersion解析MAJOR.MINOR.PATCH，忽略-及之后的后缀（如 1.2.3-rc1）
func parseVersion(version string) (major, minor, patch int, err error) {
	numbers, _, _ := strings.Cut(version, "-")
	majStr, rest, ok := strings.Cut(numbers, ".")
	if !ok || majStr == "" || rest == "" {
		return 0, 0, 0, fmt.Errorf("无效的版本格式[%s]", version)
	}
	minStr, patchStr, ok := strings.Cut(rest, ".")
	if !ok || minStr == "" || patchStr == "" {
		return 0, 0, 0, fmt.Errorf("无效的版本格式[%s]", version)
	}
	patchStr, _, _ = strings.Cut(patchStr, ".")

	if major, err = strconv.Atoi(majStr); err != nil {
		return 0, 0, 0, fmt.Errorf("无法解析major[%s]: %s", majStr, version)
	}
	if minor, err = strconv.Atoi(minStr); err != nil {
		return 0, 0, 0, fmt.Errorf("无法解析minor[%s]: %s", minStr, version)
	}
	if patch, err = strconv.Atoi(patchStr); err != nil {
		return 0, 0, 0, fmt.Errorf("无法解析patch[%s]: %s", patchStr, version)
	}
	return major, minor, patch, nil
}

// 递增版本号
func incrementVersion(version string) string {
	major, minor, patch, err := parseVersion(version)
	if err != nil {
		log.Fatalf("❌ %v", err)
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
	log.Printf("✅ 版本递增成功[%s->%s]", version, newVersion)
	return newVersion
}

// 创建并推送Git标签（伪原子操作：push失败则回滚本地tag）
func createAndPushGitTag(version string) {
	tag := tagPrefix + version

	createCmd := exec.Command("git", "tag", tag)
	var stdout, stderr bytes.Buffer
	createCmd.Stdout = &stdout
	createCmd.Stderr = &stderr
	createCmd.Dir = workdir
	if err := createCmd.Run(); err != nil {
		log.Fatalf("❌ 创建Git标签[%s]失败: %s", tag, stderr.String())
	}
	log.Printf("✅ 已创建Git标签[%s]", tag)

	pushCmd := exec.Command("git", "push", "origin", tag)
	pushCmd.Stdout = &stdout
	pushCmd.Stderr = &stderr
	pushCmd.Dir = workdir
	if err := pushCmd.Run(); err != nil {
		log.Printf("❌ 推送Git标签[%s]失败: %s", tag, stderr.String())
		rollbackCmd := exec.Command("git", "tag", "-d", tag)
		rollbackCmd.Stdout = &stdout
		rollbackCmd.Stderr = &stderr
		rollbackCmd.Dir = workdir
		if err := rollbackCmd.Run(); err != nil {
			log.Printf("📢 回滚本地标签[%s]失败: %v", tag, err)
			log.Fatalf("📢 请手动执行[git tag -d %s]", tag)
		} else {
			log.Fatalf("📢 已回滚本地标签[%s]", tag)
		}
	}
	log.Printf("✅ 已推送Git标签[%s]", tag)
}
