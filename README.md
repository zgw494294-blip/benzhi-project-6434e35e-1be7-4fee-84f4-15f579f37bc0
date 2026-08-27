# 档案文献脱酸修复工艺放行服务

本项目为档案馆纸质文献建立脱酸修复批次，登记文献件、代表性取样和酸碱度观测，执行脱酸工艺试验，自动分析酸碱度变化与强度保持率，处理偏差复验，经质量复核后冻结证据并签发可验真的入库放行凭据。服务采用版本化 JSON HTTP API，数据持久化到本地数据文件，并为每个批次维护连续审计摘要链。

建档会统一处理清单文本空白，按 `itemID` 和 `callNumber` 去重，并保存条目数、总页数及材质分布。样本必须包含观察者和观测时间；批次清单中每种材质至少登记一份有效样本后才能创建试验。试验报告按 `sampleID`、`reagent` 和工艺参数组合展示酸碱度变化、强度保持率的最小值、最大值、平均值、创建顺序趋势、参数偏差和整改复验闭环状态。

## 构建、运行与测试

```bash
go test ./...
go run ./cmd/server -addr=127.0.0.1:19081
```

运行有界自检（会通过真实回环请求完成建档、取样、试验、复核、冻结和凭据验真）：

```bash
go run ./cmd/server -selfcheck -addr=127.0.0.1:19081
```

默认监听 `127.0.0.1:19081`，也可使用 `-addr=127.0.0.1:<port>` 或 `PORT` 环境变量配置端口。除建档外，所有写入都必须用 `If-Match` 携带当前批次版本。主要路由包括 `POST /v1/batches`、`POST /v1/batches/{id}/samples`、`POST /v1/batches/{id}/trials`、`POST /v1/batches/{id}/retests`、`POST /v1/batches/{id}/review`、`POST /v1/batches/{id}/freeze`，以及 `GET /v1/batches/{id}`、`summary`、`reports`、`release`、`verify?code=` 和 `events`。冻结会在同一事务中重新计算证据摘要、验证事件序号与前序摘要、保存凭据并追加 `release.frozen` 事件；冻结后的业务写入统一返回 `FROZEN`。
