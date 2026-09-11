# 样品与生产追踪系统（Sample & Production Tracker）

包含两个追踪域：

1. **样品追踪**：样品从**收样 → 流转（设备/冰箱之间移交、上机、入库、销毁）→ 检测项目与结果登记 → 出报告**。每次流转记录「**谁**在**什么时候**把样品**从哪**移到了**哪**」。
2. **晶圆批次生产追踪**：一道工艺下有多台机台，记录每个**批次在哪台机台加工、加工起止时间、操作员、结果**，并管理**机台状态**（空闲/加工中/维护/故障）与状态变更记录。

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
- 样品位置：冰箱A2-3层、超低温冰箱B1、HPLC-01、GC-MS-02、收样台、销毁点等
- 工艺：清洗 / 光刻 / 刻蚀 / 薄膜沉积 / CMP，每道工艺下 1~3 台机台（如 ETCH-01~03）

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

### 3. 晶圆批次加工追踪（机台选择与状态记录）

**批次管理 → 批次详情**：

- 新增批次：批次号留空自动生成 `LOT-YYYYMMDD-NNN`（同样按天原子自增），填产品型号、晶圆数量。
- 「开始加工」：先选**工艺**，再从该工艺下的机台中选一台——只有**空闲**机台可选（加工中/维护/故障自动禁用并显示原因），选操作员后开始。
- 系统在同一事务内：写一条**加工记录**（机台、工艺、操作员、开始时间）、机台置为「加工中」并记录当前批次、批次进入该工艺、写一条机台状态变更日志。
- 「结束加工」：登记结果（合格/不合格/返工）、产出片数，记录结束时间并**自动算加工时长**；机台恢复空闲并写状态日志。
- 「批次完工」：所有工艺结束后标记批次已完工。加工记录表完整回答「这批晶圆在哪台机台、什么时间、谁加工的」。

**机台看板**：

- 按工艺分组展示全部机台，顶部统计空闲/加工中/维护/故障数量。
- 可登记维护、报故障、恢复空闲；加工中的机台不允许手动改状态（须在批次上结束加工）。
- 每台机台可查看**状态变更时间线**（状态、操作员、原因、时间、关联批次）。

> 并发安全：开始加工对批次和机台都加了 `SELECT ... FOR UPDATE` 行锁，同一批次不能同时开两条加工、同一机台不会被两个批次同时占用。

### 4. Wafer Bin Map 上传与缺陷分布可视化

**批次详情 →「晶圆 Map」**：批次创建时按晶圆数量自动生成槽位（#1…#N）。

- 选择槽位 → 上传该片的 **bin map 文件**，后端解析网格、bin 定义、缺口方向，并计算晶粒总数/合格数/良率/各 bin 数量。
- **Canvas 可视化**：圆形晶圆 + 缺口、晶粒按 bin 上色、鼠标悬浮显示 `(行,列) · bin`、点击图例可隐藏/显示某个 bin、右侧展示良率与 bin 汇总（缺陷集中区域一眼可见）。
- 同一晶圆多次上传保留**历史版本**，可切换查看并下载原始文件；原始文件持久化到 `MAP_STORAGE_DIR`（容器内 `/data/maps`，compose 已挂卷）。
- 单文件上限 16MB，网格上限 1000×1000。

**Bin map 文件格式**（纯文本，`backend/examples/wafer_map_example.txt` 有可直接上传的样例）：

```text
# 以 # 或 // 开头为注释
Notch: down                 # 缺口方向 up/down/left/right，可省略（默认 down）
BinDef: 1,Pass,#2ecc71,pass  # bin号,名称,#颜色,pass|fail
BinDef: 2,Edge fail,#f39c12,fail
DieSize: 5000,5000          # 可选，仅记录
Grid:
. . 1 1 2 .
. 1 1 1 2 .
. 1 1 3 2 .
```

- 也可以**只给网格**（逗号或空格分隔的 CSV）：整数是 bin 号，`.` `-` `x` `na` 为空晶粒，未声明的 bin 自动配色（约定 bin 1 为合格）。

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
| GET | `/api/v1/processes` | 工艺列表 |
| GET | `/api/v1/machines?process_id=` | 机台列表（含状态、当前批次） |
| PUT | `/api/v1/machines/:id/status` | 机台状态登记（idle/maintenance/fault） |
| GET | `/api/v1/machines/:id/status-logs` | 机台状态变更记录 |
| GET | `/api/v1/lots?keyword=&status=&page=&page_size=` | 批次分页列表 |
| POST | `/api/v1/lots` | 新增批次（批次号可自动生成） |
| GET | `/api/v1/lots/:id` | 批次详情（含全部加工记录） |
| POST | `/api/v1/lots/:id/start` | 开始加工（选机台+操作员） |
| POST | `/api/v1/lots/:id/end` | 结束加工（结果/产出，自动算时长） |
| POST | `/api/v1/lots/:id/complete` | 批次完工 |
| GET | `/api/v1/lots/:id/wafers` | 批次下的晶圆槽位（含当前 map 汇总） |
| GET | `/api/v1/wafers/:id/maps` | 单片晶圆的 map 版本列表 |
| POST | `/api/v1/wafers/:id/maps` | 上传 bin map（multipart：file/operator_id/remark） |
| GET | `/api/v1/maps/:id` | map 详情（含网格 JSON、bin 定义、汇总） |
| GET | `/api/v1/maps/:id/download` | 下载原始文件 |

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

# 新增晶圆批次（批次号自动生成）
curl -X POST http://localhost:8080/api/v1/lots \
  -H 'Content-Type: application/json' \
  -d '{"product":"LOGIC-28NM","wafer_count":25}'

# 批次 1 在 ETCH-01（种子数据机台 ID 为 10）开始加工
curl -X POST http://localhost:8080/api/v1/lots/1/start \
  -H 'Content-Type: application/json' \
  -d '{"machine_id":10,"operator_id":1,"remark":"刻蚀工艺首件"}'

# 结束加工：合格、产出 25 片（自动算时长，机台恢复空闲）
curl -X POST http://localhost:8080/api/v1/lots/1/end \
  -H 'Content-Type: application/json' \
  -d '{"result":"ok","wafer_out":25,"operator_id":1}'

# 机台登记维护
curl -X PUT http://localhost:8080/api/v1/machines/4/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"maintenance","operator_id":2,"reason":"定期保养"}'

# 给 1 号晶圆上传 bin map
curl -X POST http://localhost:8080/api/v1/wafers/1/maps \
  -F operator_id=1 -F remark="刻蚀后量测" \
  -F file=@backend/examples/wafer_map_example.txt
```

## 数据模型

```
# 样品域
users          操作员（两个域共用）
locations      位置（设备 device / 冰箱 fridge / 实验台 bench / 销毁点 discard）
samples        样品（唯一编号、当前位置、状态）
transfers      流转记录（from/to location, operator, action, occurred_at）
test_results   检测项目与结果
daily_seqs     样品编号按天自增序列

# 生产域
processes            工艺（一道工艺多台机台）
machines             机台（归属工艺、状态、当前加工批次）
wafer_lots           晶圆批次（批次号、片数、状态、当前工艺）
wafers               晶圆（批次槽位 1..N、当前 map）
wafer_bin_maps       bin map（JSONB 网格、bin 定义、良率/bin 汇总、版本、原始文件路径）
lot_daily_seqs       批次号按天自增序列
lot_process_records  加工记录（批次×工艺×机台×操作员，开始/结束时间、时长、结果、产出）
machine_status_logs  机台状态变更记录
```

样品/批次与位置、机台、记录均为外键关联；样品、工艺、机台、批次软删除（GORM DeletedAt），加工与流转记录保留作为审计轨迹。

## 目录结构

```
.
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── model/       # GORM 模型（含 JSONB map 类型）
│   │   ├── binmap/      # bin map 文本解析器（含单元测试）
│   │   ├── dto/         # 请求/响应结构
│   │   ├── database/    # 连接、迁移、种子数据
│   │   ├── service/     # 业务逻辑（编号生成、事务、map 解析汇总）
│   │   ├── handler/     # Gin 控制器
│   │   └── router/
│   ├── examples/        # 可直接上传的 wafer map 样例
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── api/         # axios 封装与接口
│   │   ├── stores/      # Pinia: sample / meta / production
│   │   ├── types/       # 与后端对应的 TS 类型
│   │   ├── views/       # 批次列表/详情、机台看板、晶圆 Map、样品列表/详情
│   │   └── components/  # 加工/流转/上传对话框、WaferMapCanvas 渲染组件
│   └── Dockerfile + nginx.conf
└── docker-compose.yml
```

## 后续可扩展

- 登录鉴权与操作权限（当前操作员为下拉选择，便于演示）
- 位置层级（冰箱 → 层架 → 冻存盒）与条码/二维码扫码
- 检测方法库、报告 PDF 导出、温度异常告警
- 流转事件的完整审计日志（含修改/删除留痕）
- 工艺路线（按产品配置工序顺序与必过工艺）、批次跳站/返工流转校验
- 机台 OEE 统计、加工节拍/时长分析、与 MES/EAP 设备联机自动采集状态
- Wafer map：对接 STDF / KLA / KLARF 等量测格式、缺陷坐标散点叠加、跨片/跨批缺陷模式聚类与 SPC 报警、map 文件对象存储（S3/MinIO）
