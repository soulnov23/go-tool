FROM centos:latest

COPY ./build/bin /app/bin
COPY ./build/conf /app/conf

RUN chmod +x /app/bin/go-tool

ENTRYPOINT ["/bin/bash", "-c", "ulimit -c unlimited && export GOTRACEBACK=crash && cd /app/bin && ./go-tool -conf ../conf/go_tool.yaml"]
