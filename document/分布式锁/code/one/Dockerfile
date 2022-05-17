FROM  golang:1.15-alpine as builder

# 启用 Go Modules 功能 go 设置国内源
RUN go env -w GO111MODULE=on \
    && go env -w GOSUMDB=off \
    && go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# cd 到 /workspace 里面
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum

# 复制库
COPY main.go/ main.go/
# 执行 go build   从./cmd/user/server.go  输出到 ./app 可执行文件
# 最终 app 可执行文件 路径 为  /workspace/app
RUN CGO_ENABLED=0   go build -a -o app ./main.go

EXPOSE 8050
# 多段编译
FROM  alpine
# 修改时区
# 修改时区
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
# cd 到 / 目录
WORKDIR /
# 把builder 阶段 的 成功 /workspace/app 可执行文件 复制到 当前目录来 (/workspace)
COPY --from=builder /workspace/app .
## 暴露容器端口 8050
EXPOSE 8050
#执行  /app 命令  运行 app 启动服务

ENTRYPOINT ["/app"]
#CMD ["/one"]
