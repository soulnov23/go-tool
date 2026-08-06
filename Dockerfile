FROM centos:latest

RUN yum update -y && yum -y install vim grep iputils net-tools telnet procps-ng htop lsof curl wget tcpdump strace lrzsz && yum clean all

COPY ./build/bin/ /app/bin/
COPY ./build/conf/ /app/conf/

RUN chmod +x /app/bin/*

ENTRYPOINT ["/bin/bash", "-c", "ulimit -c unlimited && export GOTRACEBACK=crash && cd /app/bin && ./go-tool -conf /app/conf/go_tool.yaml"]
