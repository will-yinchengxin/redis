# Redis 高级数据类型实战教程（HyperLogLog + Geospatial + Bitmap）

这是一个完整的 Redis 高级数据类型学习项目，使用 Golang 实现，包含20+实战示例。

## 📚 项目内容

### 三种数据类型全覆盖

1. **HyperLogLog** (5个示例)
   - 唯一计数估计
   - 12KB 内存存储亿级数据
   - 适合：UV统计、DAU/MAU

2. **Geospatial** (8个示例)
   - 地理空间索引
   - 位置存储与范围查询
   - 适合：外卖配送、附近的人

3. **Bitmap** (7个示例)
   - 位图操作
   - 极致内存效率
   - 适合：签到、在线状态、权限

## 🚀 快速开始

### 环境准备

```bash
# 安装 Redis
# macOS
brew install redis
brew services start redis

# Ubuntu
sudo apt install redis-server
sudo systemctl start redis

# Docker
docker run -d -p 6379:6379 --name redis redis:latest
```

### 运行项目

```bash
# 1. 下载依赖
go mod download

# 2. 运行程序
go run .

# 3. 或者编译后运行
go build -o redis-demo
./redis-demo
```

### 交互式菜单

```
1. HyperLogLog 示例
2. Geospatial 示例
3. Bitmap 示例
4. 运行所有示例
5. 快速对比演示
6. 三种数据类型综合对比
```

## 📖 学习路径

### 🔰 初学者（30分钟）

```
1. 阅读 REDIS_ADVANCED_TUTORIAL.md（理解原理）
2. 运行"快速对比演示"（5分钟体验）
3. 查看代码实现（理解API用法）
```

### 🎯 进阶学习（2小时）

```
1. 阅读完整教程（30分钟）
2. 运行所有示例（30分钟）
3. 研究代码细节（30分钟）
4. 动手修改参数（30分钟）
```

### 💼 实战开发（按需）

```
1. 确定使用场景
2. 查找对应示例
3. 复制代码片段
4. 根据需求调整
```

## 🎯 示例列表

### HyperLogLog 示例

| 示例 | 说明 | 适用场景 |
|------|------|---------|
| 示例1 | 基础使用 | 学习API |
| 示例2 | 网站UV统计 | 电商、新闻网站 |
| 示例3 | 多天数据合并 | 周活、月活计算 |
| 示例4 | 性能内存对比 | 理解优势 |
| 示例5 | App DAU/MAU | 运营数据分析 |

### Geospatial 示例

| 示例 | 说明 | 适用场景 |
|------|------|---------|
| 示例1 | 基础使用 | 学习API |
| 示例2 | 查找附近位置 | 外卖、点评App |
| 示例3 | 根据成员查找 | 旅游推荐 |
| 示例4 | 网约车配送 | 滴滴、美团 |
| 示例5 | 动态更新位置 | 实时定位 |
| 示例6 | 区域查询 | 地图服务 |
| 示例7 | GeoHash编码 | 理解原理 |
| 示例8 | 删除位置 | 用户离线 |

### Bitmap 示例

| 示例 | 说明 | 适用场景 |
|------|------|---------|
| 示例1 | 基础使用 | 学习API |
| 示例2 | 每日签到 | 签到系统 |
| 示例3 | 用户活跃统计 | 留存分析 |
| 示例4 | A/B测试分组 | 实验系统 |
| 示例5 | 用户权限管理 | 权限系统 |
| 示例6 | 内存效率对比 | 理解优势 |
| 示例7 | 在线用户统计 | 实时监控 |

## 💡 核心特性

### HyperLogLog
```
✓ 固定 12KB 内存
✓ 支持亿级数据
✓ 约 1% 误差
✓ 可合并多个 HLL
```

### Geospatial
```
✓ 存储经纬度
✓ 计算距离
✓ 范围查询
✓ 基于 GeoHash
```

### Bitmap
```
✓ 1 bit 存 1 个用户
✓ 100% 精确
✓ 位运算支持
✓ 需要连续 ID
```

## 📊 性能数据

### 内存对比（100万用户）

| 方案 | 内存占用 | 精确度 |
|------|---------|--------|
| Set | ~36 MB | 100% |
| Bitmap | ~122 KB | 100% |
| HyperLogLog | 12 KB | ~99% |

### 适用场景对比

| 需求 | 推荐方案 | 原因 |
|------|---------|------|
| 统计 UV | HyperLogLog | 内存小 |
| 用户签到 | Bitmap | ID连续 |
| 附近的人 | Geospatial | 位置数据 |
| 需要列表 | Set | 可获取元素 |

## 🔍 选择指南

```
需要知道具体元素？
├─ 是 → Set
└─ 否 → 需要精确计数？
    ├─ 是 → ID连续？
    │   ├─ 是 → Bitmap
    │   └─ 否 → Set(小规模) / HLL(大规模)
    └─ 否 → HyperLogLog

位置数据？ → Geospatial
```

## 📝 代码示例

### HyperLogLog

```go
// 统计 UV
key := fmt.Sprintf("hll:uv:%s", date)
client.PFAdd(ctx, key, userID)
count, _ := client.PFCount(ctx, key).Result()

// 计算 MAU
client.PFMerge(ctx, "hll:mau", key1, key2, key3...)
mau, _ := client.PFCount(ctx, "hll:mau").Result()
```

### Bitmap

```go
// 用户签到
key := fmt.Sprintf("bitmap:signin:%d:%s", userID, month)
client.SetBit(ctx, key, int64(day-1), 1)

// 统计签到天数
count, _ := client.BitCount(ctx, key, nil).Result()

// 计算留存（位运算）
client.BitOpAnd(ctx, "retention", "day1", "day7")
```

### Geospatial

```go
// 添加位置
client.GeoAdd(ctx, "geo:riders", &redis.GeoLocation{
    Name:      "rider1",
    Longitude: 116.397428,
    Latitude:  39.909186,
})

// 查找附近 5km 的骑手
results, _ := client.GeoRadius(ctx, "geo:riders", 
    lng, lat, &redis.GeoRadiusQuery{
        Radius: 5,
        Unit:   "km",
        Count:  10,
    }).Result()
```

## 🎓 进阶学习

### 推荐阅读顺序

1. REDIS_ADVANCED_TUTORIAL.md（理论基础）
2. 运行示例程序（实践）
3. 查看源代码（实现细节）
4. 阅读最佳实践（避坑指南）

### 实战项目建议

1. **UV/PV 统计系统**（HyperLogLog）
2. **用户签到系统**（Bitmap）
3. **外卖配送系统**（Geospatial）
4. **留存率分析系统**（Bitmap位运算）

## ⚠️ 注意事项

### HyperLogLog
- ❌ 无法获取具体元素
- ❌ 不能删除元素
- ✓ 适合只需计数的场景

### Bitmap
- ❌ ID必须是连续整数
- ❌ ID间隔大会浪费空间
- ✓ 适合用户ID、天数等连续值

### Geospatial
- ❌ 单key不宜超过100万位置
- ✓ 建议按城市/区域分片
- ✓ 定期清理过期数据

## 🛠️ 常见问题

**Q: 如何选择数据结构？**
```
只要计数 → HyperLogLog
连续ID+布尔值 → Bitmap
位置数据 → Geospatial
需要元素列表 → Set
```

**Q: HyperLogLog 误差可接受吗？**
```
对于百万级数据，误差通常 < 0.5%
运营数据分析完全可接受
需要精确值请用 Bitmap 或 Set
```

**Q: Bitmap 的 ID 不连续怎么办？**
```
方案1: 建立 ID 映射表
方案2: 使用 HyperLogLog
方案3: 使用 Set（数据量小时）
```

## 📞 获取帮助

遇到问题时：
1. 查看代码注释
2. 阅读完整教程
3. 查看 Redis 官方文档

## 📄 文件说明

- `main.go` - 主程序
- `hyperloglog_demo.go` - HyperLogLog 示例
- `geospatial_demo.go` - Geospatial 示例
- `bitmap_demo.go` - Bitmap 示例
- `REDIS_ADVANCED_TUTORIAL.md` - 完整教程
- `go.mod` - 依赖管理

## 🎉 开始学习

```bash
go run .
# 选择 5 - 快速对比演示
# 5 分钟了解三种数据类型的特点
```

