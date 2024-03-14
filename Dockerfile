FROM centos:latest

COPY ./build/bin /app/bin
COPY ./build/conf /app/conf

RUN chmod +x /app/bin/*

ENTRYPOINT ["/bin/bash", "-c", "cd /app/bin && ./start.sh"]
