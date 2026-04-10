# IoT Server

基于 Go 语言的 IoT 物联网平台后端服务。

## 技术栈

- **Web 框架**: Gin
- **ORM**: GORM + MySQL
- **缓存**: Redis
- **MQTT**: EMQX (Paho MQTT Client)
- **消息队列**: Apache Kafka
- **配置管理**: Viper
- **日志**: Zap

## 项目结构

```
iot/
├── cmd/server/             # 程序入口
├── internal/               # 私有应用代码
│   ├── handler/            # HTTP 处理器
│   ├── service/            # 业务逻辑层
│   ├── repository/         # 数据访问层
│   ├── model/              # 数据模型
│   ├── middleware/          # 中间件
│   ├── router/             # 路由
│   └── mq/                 # 消息队列
│       ├── emqx/           # EMQX MQTT
│       └── kafka/          # Kafka
├── pkg/                    # 公共库
│   ├── config/             # 配置
│   ├── database/           # 数据库
│   ├── redis/              # Redis
│   ├── logger/             # 日志
│   └── response/           # 统一响应
├── config/                 # 配置文件
├── deploy/                 # 部署配置
├── Makefile
└── README.md
```

## 快速开始

### 1. 启动依赖服务

```bash
make docker-up
```

### 2. 修改配置

编辑 `config/config.yaml`，根据实际环境调整数据库、Redis、Kafka、EMQX 的连接信息。

### 3. 安装依赖 & 运行

```bash
make tidy
make run
```

### 4. 编译

```bash
make build
```

## API 示例

```bash
# 创建设备
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -d '{"device_name":"温度传感器","device_key":"temp-001","product_key":"sensor","status":0}'

# 获取设备列表
curl http://localhost:8080/api/v1/devices

# 获取设备详情
curl http://localhost:8080/api/v1/devices/1

# 健康检查
curl http://localhost:8080/health
```

## 端口说明

| 服务 | 端口 |
|------|------|
| HTTP Server | 8080 |
| MySQL | 3306 |
| Redis | 6379 |
| Kafka | 9092 |
| Zookeeper | 2181 |
| EMQX MQTT | 1883 |
| EMQX WebSocket | 8083 |
| EMQX Dashboard | 18083 |
