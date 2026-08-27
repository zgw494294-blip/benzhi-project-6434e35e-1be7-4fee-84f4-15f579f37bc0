# BENZHI_README

基于 Go 实现的档案文献脱酸修复工艺放行服务 HTTP API 项目，一款后端服务，档案文献脱酸修复工艺放行服务提供批次建档、代表性取样、脱酸试验、指标分析、偏差复验、质量复核、证据冻结和放行凭据验真能力。

## 项目说明
- 项目：benzhi-project-6434e35e-1be7-4fee-84f4-15f579f37bc0
- 项目用途：档案文献脱酸修复工艺放行服务提供批次建档、代表性取样、脱酸试验、指标分析、偏差复验、质量复核、证据冻结和放行凭据验真能力。
- Go 工具链：`golang:1.23`
- 前端工具链：无

## 标准构建、运行和测试命令
进入容器后执行：
cd '/app' && GOTOOLCHAIN=local go build ./...
cd /app && GOTOOLCHAIN=local go run ./cmd/server -selfcheck -addr=127.0.0.1:19081
cd '/app' && GOTOOLCHAIN=local go test ./...

## Docker 构建和进入容器
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-project-6434e35e-1be7-4fee-84f4-15f579f37bc0-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-project-6434e35e-1be7-4fee-84f4-15f579f37bc0-arm64 linux/arm64
docker run -it benzhi-project-6434e35e-1be7-4fee-84f4-15f579f37bc0-amd64:latest

## 题目验证命令
1. 预期退出码 0：`go test ./...`
2. 预期退出码 0：`go run ./cmd/server -selfcheck -addr=127.0.0.1:19081`
