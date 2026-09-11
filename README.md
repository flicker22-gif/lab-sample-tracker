# 样品全流程追踪系统（Sample Tracker）

覆盖样品从**收样 → 流转（设备/冰箱之间移交、上机、入库、销毁）→ 检测项目与结果登记 → 出报告**的全流程追踪。
每次流转都记录「**谁**在**什么时候**把样品**从哪**移到了**哪**」，形成完整轨迹。

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3 + TypeScript + Vite + Pinia + Vue Router + Element Plus + Axios |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 15 |
| 部署 | Docker Compose（前端 Nginx 托管 + API 反代） |

## 一键启动（推荐）

需要本机有 Docker / Docker Compose。

```bash
docker compose up -d --build
```

- 前端：http://localhost:8088
- 后端 API：http://localhost:8080
- 健康检查：http://localhost:8080/healthz

首次启动自动建表（GORM AutoMigrate）并写入演示数据：

- 操作员：张三 / 李四 / 王五
- 位置：冰箱A2-3层、超低温冰箱B1、HPLC-01、GC-MS-02、收样台、销毁点等

## 核心功能

### 1. 新增样品（收样登记）

样品列表 → 「新增样品」：填写名称、类型、送检单位、收样人、首次存放位置。

- 样品编号由系统自动生成，格式 `SP-YYYYMMDD-NNNN`（如 `SP-20260910-0001`），按天自增。
- 编号通过 PostgreSQL `INSERT ... ON CONFLICT DO UPDATE` 原子递增，**并发收样不重号**。
- 收样与第一条 `receive` 流转记录在同一个数据库事务中创建。

### 2. 查看流转轨迹

点击样品编号进入详情页：

- 左侧**时间线**按时间倒序展示全部流转节点：动作（收样/移交/上机/下机/入库/销毁）、操作人、来源位置 → 目标位置、时间、备注。
- 右侧展示该样品关联的**检测项目、结果值、单位、结论、报告编号、检测人**。
- 「登记流转」可记录一次位置变化，样品当前位置与状态（检测中/存储中/已出报告/已销毁）随之更新。
- 「登记检测结果」填写报告编号后，样品自动标记为「已出报告」。

## 本地开发

### 后端

```bash
cd backend
# 先准备好 PostgreSQL 15（可用 compose 只起数据库）：
#   docker compose up -d postgres
export DB_HOST=localhost DB_USER=tracker DB_PASSWORD=tracker DB_NAME=tracker
go mod tidy
go run ./cmd/server          # 监听 :8080
```

国内网络可设置 `GOPROXY=https://goproxy.cn,direct`。

### 前端

```bash
cd frontend
npm install
npm run dev                  # http://localhost:5173，/api 已代理到 :8080
```

类型检查 + 构建：

```bash
npm run build
```

## API 一览

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/samples?keyword=&status=&page=&page_size=` | 样品分页列表 |
| POST | `/api/v1/samples` | 新增样品（自动编号 + 收样记录） |
| GET | `/api/v1/samples/:id` | 样品详情（含流转轨迹、检测结果） |
| POST | `/api/v1/samples/:id/transfers` | 登记一次流转 |
| POST | `/api/v1/samples/:id/results` | 登记检测项目/结果 |
| GET | `/api/v1/users` | 操作员列表 |
| GET | `/api/v1/locations?type=` | 位置列表（device/fridge/bench/discard） |

统一响应：`{ "code": 0, "message": "ok", "data": ... }`。

### 请求示例

```bash
# 新增样品
curl -X POST http://localhost:8080/api/v1/samples \
  -H 'Content-Type: application/json' \
  -d '{"name":"出厂水","category":"水质","source":"某水厂","receiver_id":1,"location_id":1}'

# 登记流转：从冰箱移到 HPLC-01 上机
curl -X POST http://localhost:8080/api/v1/samples/1/transfers \
  -H 'Content-Type: application/json' \
  -d '{"to_location_id":5,"operator_id":2,"action":"load","note":"开始铅项目检测"}'

# 登记检测结果（带报告编号 → 样品标记已出报告）
curl -X POST http://localhost:8080/api/v1/samples/1/results \
  -H 'Content-Type: application/json' \
  -d '{"item_name":"铅(Pb)","method":"GB 5749-2022","instrument":"ICP-MS-01","result_value":"0.003","unit":"mg/L","conclusion":"qualified","report_no":"BG-20260910-01","analyst_id":2}'
```

## 数据模型

```
users          操作员
locations      位置（设备 device / 冰箱 fridge / 实验台 bench / 销毁点 discard）
samples        样品（唯一编号、当前位置 current_location_id、状态）
transfers      流转记录（sample_id, from_location_id, to_location_id, operator_id, action, occurred_at）
test_results   检测项目与结果（项目、方法、仪器、结果值、结论、报告编号、检测人）
daily_seqs     样品编号按天自增序列
```

样品与位置、流转、结果均为外键关联；样品软删除（GORM DeletedAt），流转记录保留作为审计轨迹。

## 目录结构

```
.
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── model/       # GORM 模型
│   │   ├── dto/         # 请求/响应结构
│   │   ├── database/    # 连接、迁移、种子数据
│   │   ├── service/     # 业务逻辑（编号生成、事务）
│   │   ├── handler/     # Gin 控制器
│   │   └── router/
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── api/         # axios 封装与接口
│   │   ├── stores/      # Pinia: sample / meta
│   │   ├── types/       # 与后端对应的 TS 类型
│   │   ├── views/       # 样品列表、样品详情（轨迹）
│   │   └── components/  # 新增样品/流转登记/结果登记对话框
│   └── Dockerfile + nginx.conf
└── docker-compose.yml
```

## 后续可扩展

- 登录鉴权与操作权限（当前操作员为下拉选择，便于演示）
- 位置层级（冰箱 → 层架 → 冻存盒）与条码/二维码扫码
- 检测方法库、报告 PDF 导出、温度异常告警
- 流转事件的完整审计日志（含修改/删除留痕）
