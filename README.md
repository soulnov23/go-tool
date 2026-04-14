- [go-tool](#go-tool)
  - [📖 简介](#-简介)
    - [Go Reference](#go-reference)
    - [工程规范](#工程规范)
    - [编码规范](#编码规范)
  - [🚀 功能](#-功能)
    - [debug](#debug)
  - [💡 流程](#-流程)
  - [🧰 安装](#-安装)
  - [⚙️ 设置](#️-设置)
  - [🧲 效果](#-效果)
  - [📚 链接](#-链接)

# go-tool

```
                                                               tttt                                            lllllll 
                                                            ttt:::t                                            l:::::l 
                                                            t:::::t                                            l:::::l 
                                                            t:::::t                                            l:::::l 
   ggggggggg   ggggg   ooooooooooo                    ttttttt:::::ttttttt       ooooooooooo      ooooooooooo    l::::l 
  g:::::::::ggg::::g oo:::::::::::oo                  t:::::::::::::::::t     oo:::::::::::oo  oo:::::::::::oo  l::::l 
 g:::::::::::::::::go:::::::::::::::o                 t:::::::::::::::::t    o:::::::::::::::oo:::::::::::::::o l::::l 
g::::::ggggg::::::ggo:::::ooooo:::::o --------------- tttttt:::::::tttttt    o:::::ooooo:::::oo:::::ooooo:::::o l::::l 
g:::::g     g:::::g o::::o     o::::o -:::::::::::::-       t:::::t          o::::o     o::::oo::::o     o::::o l::::l 
g:::::g     g:::::g o::::o     o::::o ---------------       t:::::t          o::::o     o::::oo::::o     o::::o l::::l 
g:::::g     g:::::g o::::o     o::::o                       t:::::t          o::::o     o::::oo::::o     o::::o l::::l 
g::::::g    g:::::g o::::o     o::::o                       t:::::t    tttttto::::o     o::::oo::::o     o::::o l::::l 
g:::::::ggggg:::::g o:::::ooooo:::::o                       t::::::tttt:::::to:::::ooooo:::::oo:::::ooooo:::::ol::::::l
 g::::::::::::::::g o:::::::::::::::o                       tt::::::::::::::to:::::::::::::::oo:::::::::::::::ol::::::l
  gg::::::::::::::g  oo:::::::::::oo                          tt:::::::::::tt oo:::::::::::oo  oo:::::::::::oo l::::::l
    gggggggg::::::g    ooooooooooo                              ttttttttttt     ooooooooooo      ooooooooooo   llllllll
            g:::::g                                                                                                    
gggggg      g:::::g                                                                                                    
g:::::gg   gg:::::g                                                                                                    
 g::::::ggg:::::::g                                                                                                    
  gg:::::::::::::g                                                                                                     
    ggg::::::ggg                                                                                                       
       gggggg                                                                                                          
```

generate by http://patorjk.com/software/taag/#p=display&f=Doh&t=go-tool

## 📖 简介

### Go Reference 

[![Go Reference](https://pkg.go.dev/badge/github.com/soulnov23/go-tool.svg)](https://pkg.go.dev/github.com/soulnov23/go-tool)

### 工程规范

[https://github.com/golang-standards/project-layout/blob/master/README_zh.md](https://github.com/golang-standards/project-layout/blob/master/README_zh.md)

### 编码规范

[https://github.com/golang/go/wiki/CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)

## 🚀 功能

### debug

```shell
curl -v 'http://127.0.0.1:6060/debug/pprof/profile?seconds=30' > profile.tar.gz
curl -v 'http://127.0.0.1:6060/debug/pprof/heap?seconds=30' > heap.tar.gz
curl -v 'http://127.0.0.1:6060/debug/pprof/goroutine?seconds=30' > goroutine.tar.gz

yum install -y graph
go tool pprof -http 0.0.0.0:9999 profile.tar.gz
go tool pprof -http 0.0.0.0:9999 heap.tar.gz
go tool pprof -http 0.0.0.0:9999 goroutine.tar.gz

go tool pprof -http 0.0.0.0:9999 'http://127.0.0.1:6060/debug/pprof/profile?seconds=30'
go tool pprof -http 0.0.0.0:9999 'http://127.0.0.1:6060/debug/pprof/heap?seconds=30'
go tool pprof -http 0.0.0.0:9999 'http://127.0.0.1:6060/debug/pprof/goroutine?seconds=30'
```

- debug=0 默认值，返回二进制的Protocol Buffers格式文件，这是go tool pprof工具期望的格式，适合用于后续的专业分析，但人类无法直接阅读
- debug=1 纯文本格式，其中包含符号化的函数名、源文件路径和行号等信息，可以直接在浏览器中阅读，适合快速查看
- debug=2 对于/debug/pprof/goroutine，此参数会以极高的可读性输出每个goroutine的完整状态和堆栈跟踪，格式类似于程序发生panic时的堆栈输出，非常适合深度调试goroutine泄漏或阻塞问题

## 💡 流程
## 🧰 安装
## ⚙️ 设置
## 🧲 效果
## 📚 链接