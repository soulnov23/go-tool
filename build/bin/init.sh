#!/bin/bash

ulimit -c unlimited
export GOTRACEBACK=crash
#allocfreetrace输出内存分配和释放的详细跟踪信息
#gctrace输出垃圾回收的详细跟踪信息
#inittrace输出init函数的执行顺序
#export GODEBUG=allocfreetrace=1,gctrace=1,inittrace=1