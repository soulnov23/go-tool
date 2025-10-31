#-v编译包时打印包的名称
#-x打印命令
PRINT := -v -x

#使用go tool compile查看-gcflags传递给编译器的参数
#-N禁用优化
#-l禁用内联
#使用go tool compile -d help查看调试设置
#-d=checkptr检查unsafe.Pointer转换
#0:禁用检查
#1:检查unsafe.Pointer转换
#0:转换为unsafe.Pointer的对象分配到堆上
DEBUG := -gcflags "all=-N -l -d=checkptr=1"

#使用go tool link查看-ldflags传递给链接器的参数
#-w禁用DWARF生成
#-s禁用符号表
RELEASE := -ldflags "-w -s"

#-m打印优化策略，编译器优化技术确定变量是否需要在堆上分配内存
ESCAPE := -gcflags "-m"

VERSION := -ldflags "-X 'main.goVersion=$(shell go version)' \
					 -X 'main.gitBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
					 -X 'main.gitCommitID=$(shell git rev-parse HEAD)' \
					 -X 'main.gitCommitTime=$(shell git log --pretty=format:"%ci" | head -1)' \
					 -X 'main.gitCommitAuthor=$(shell git log --pretty=format:"%cn" | head -1)'"

CGO := CGO_ENABLED=0