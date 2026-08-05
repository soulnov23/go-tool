package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

var drivers = map[string]string{
	".c":   "clang",
	".cc":  "clang++",
	".cpp": "clang++",
	".cxx": "clang++",
}

type compileCommand struct {
	Directory string   `json:"directory"`
	Arguments []string `json:"arguments"`
	File      string   `json:"file"`
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("\033[1;32m[clangd-launch]\033[m ")

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ 获取当前工作目录失败: %v", err)
	}

	// 刷新compile_commands.json
	if err := refreshCompileCommands(cwd); err != nil {
		log.Fatalf("❌ 刷新compile_commands.json失败: %v", err)
	}
}

// 扫描源文件并按需刷新compile_commands.json
func refreshCompileCommands(root string) error {
	var sources []string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return err
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if _, ok := drivers[ext]; !ok {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sources = append(sources, rel)
		return nil
	}); err != nil {
		return err
	}
	if len(sources) == 0 {
		log.Print("📢 未扫描到源文件，跳过刷新compile_commands.json")
		return nil
	}

	commands := make([]compileCommand, 0, len(sources))
	for _, src := range sources {
		ext := filepath.Ext(src)
		commands = append(commands, compileCommand{
			Directory: root,
			Arguments: []string{drivers[ext], "-c", src},
			File:      src,
		})
	}

	payload, err := json.MarshalIndent(commands, "", "    ")
	if err != nil {
		return err
	}

	output := filepath.Join(root, "compile_commands.json")
	if existing, err := os.ReadFile(output); err == nil && bytes.Equal(existing, payload) {
		log.Printf("📢 compile_commands.json已是最新[%d个源文件]", len(sources))
		return nil
	}
	if err := os.WriteFile(output, payload, 0o644); err != nil {
		return err
	}
	log.Printf("✅ 已刷新compile_commands.json[%d个源文件]", len(sources))
	return nil
}
