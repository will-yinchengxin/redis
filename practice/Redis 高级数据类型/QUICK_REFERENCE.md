# Redis 高级数据类型速查手册

## 一、三种数据类型对比

| 数据类型 | 内存占用 | 精确度 | 适用场景 | ID要求 |
|---------|---------|--------|---------|--------|
| **HyperLogLog** | 固定12KB | ~99% | 大规模计数 | 任意 |
| **Bitmap** | ID最大值/8 | 100% | 布尔值记录 | 连续整数 |
| **Geospatial** | ~50字节/位置 | 100% | 位置服务 | 无 |

## 二、快速选择指南

```
┌─ 只要计数，不需要元素？
│  └─ HyperLogLog
│
├─ 连续ID + 布尔值？
│  └─ Bitmap
│
├─ 位置数据？
│  └─ Geospatial
│
└─ 需要元素列表？
   └─ Set
```

## 三、常用命令

### HyperLogLog

```go
// 添加
client.PFAdd(ctx, "hll:uv", "user1", "user2")

// 计数
count, _ := client.PFCount(ctx, "hll:uv").Result()

// 合并
client.PFMerge(ctx, "hll:mau", "hll:day1", "hll:day2")
```

### Bitmap

```go
// 设置
client.SetBit(ctx, "bitmap:signin", 10, 1)

// 获取
val, _ := client.GetBit(ctx, "bitmap:signin", 10).Result()

// 统计
count, _ := client.BitCount(ctx, "bitmap:signin", nil).Result()

// 位运算
client.BitOpAnd(ctx, "result", "key1", "key2")
```

### Geospatial

```go
// 添加位置
client.GeoAdd(ctx, "geo:riders", &redis.GeoLocation{
    Name: "rider1", Longitude: 116.4, Latitude: 39.9,
})

// 计算距离
dist, _ := client.GeoDist(ctx, "geo:riders", "r1", "r2", "km").Result()

// 范围查询
results, _ := client.GeoRadius(ctx, "geo:riders", 116.4, 39.9,
    &redis.GeoRadiusQuery{Radius: 5, Unit: "km"}).Result()
```

## 四、典型应用场景

### 网站UV统计（HyperLogLog）

```go
// 记录访问
date := time.Now().Format("2006-01-02")
key := fmt.Sprintf("hll:uv:%s", date)
client.PFAdd(ctx, key, userID)
client.Expire(ctx, key, 90*24*time.Hour)

// 获取UV
uv, _ := client.PFCount(ctx, key).Result()

// 计算MAU
keys := []string{"hll:uv:2024-01-01", "hll:uv:2024-01-02", ...}
client.PFMerge(ctx, "hll:mau", keys...)
mau, _ := client.PFCount(ctx, "hll:mau").Result()
```

### 用户签到（Bitmap）

```go
// 签到
userID := 12345
month := "2024-01"
day := 15
key := fmt.Sprintf("signin:%d:%s", userID, month)
client.SetBit(ctx, key, int64(day-1), 1)

// 查询是否签到
signed, _ := client.GetBit(ctx, key, int64(day-1)).Result()

// 统计签到天数
count, _ := client.BitCount(ctx, key, nil).Result()

// 计算连续签到
// 从今天往前遍历，遇到0就停止
```

### 外卖派单（Geospatial）

```go
// 骑手上线
client.GeoAdd(ctx, "geo:riders:beijing", &redis.GeoLocation{
    Name: "rider001", Longitude: 116.4, Latitude: 39.9,
})

// 查找最近的3个骑手
results, _ := client.GeoRadius(ctx, "geo:riders:beijing",
    orderLng, orderLat, &redis.GeoRadiusQuery{
        Radius: 5,
        Unit:   "km",
        Count:  3,
        Sort:   "ASC",
        WithDist: true,
    }).Result()

// 派单给最近的骑手
selectedRider := results[0].Name

// 骑手接单后移除
client.ZRem(ctx, "geo:riders:beijing", selectedRider)
```

### 活跃用户分析（Bitmap）

```go
// 记录每日活跃用户
date := "2024-01-15"
key := fmt.Sprintf("bitmap:active:%s", date)
client.SetBit(ctx, key, int64(userID), 1)

// 计算7日留存（7天都活跃的用户）
keys := []string{
    "bitmap:active:2024-01-15",
    "bitmap:active:2024-01-16",
    // ... 共7天
}
client.BitOpAnd(ctx, "bitmap:retention:7day", keys...)
retention, _ := client.BitCount(ctx, "bitmap:retention:7day", nil).Result()

// 计算至少活跃1天的用户
client.BitOpOr(ctx, "bitmap:any:7day", keys...)
anyActive, _ := client.BitCount(ctx, "bitmap:any:7day", nil).Result()

// 留存率 = retention / anyActive
```

## 五、性能优化

### HyperLogLog

```go
// ✓ 按日期分key
key := fmt.Sprintf("hll:uv:%s", date)

// ✓ 设置过期时间
client.Expire(ctx, key, 90*24*time.Hour)

// ✓ 批量添加
client.PFAdd(ctx, key, user1, user2, user3, ...)

// ✗ 不要所有数据放一个key
```

### Bitmap

```go
// ✓ 确保ID连续
// ID: 1, 2, 3, 4, 5 ... ✓
// ID: 1, 100, 1000, 10000 ... ✗ (浪费空间)

// ✓ 使用位运算分析
client.BitOpAnd(ctx, "result", "day1", "day7")

// ✓ 范围统计
client.BitCount(ctx, key, &redis.BitCount{
    Start: 0,
    End:   30, // 只统计前31天
})
```

### Geospatial

```go
// ✓ 按城市分片
key := fmt.Sprintf("geo:%s:riders", city)

// ✓ 限制搜索半径
query.Radius = 5  // 不要太大

// ✓ 限制返回数量
query.Count = 20

// ✓ 定期清理过期数据
// 删除1小时未更新的位置
```

## 六、常见错误

### ❌ 错误1: HyperLogLog用于小数据

```go
// 1000个用户用HyperLogLog → 没必要，直接用Set
// HyperLogLog适合百万、千万级数据
```

### ❌ 错误2: Bitmap用于非连续ID

```go
// UUID映射到Bitmap → 无法使用
// 应该用HyperLogLog或Set
```

### ❌ 错误3: Geospatial单key过多数据

```go
// 一个key存100万位置 → 性能下降
// 应该按城市/区域分片
```

### ❌ 错误4: 忘记设置过期时间

```go
// HyperLogLog和Bitmap都应该设置TTL
client.Expire(ctx, key, 90*24*time.Hour)
```

## 七、选择决策

### 场景：统计UV

```
用户ID: UUID → HyperLogLog
用户ID: 1,2,3... → Bitmap或HyperLogLog
需要精确值 → Bitmap
可接受误差 → HyperLogLog
```

### 场景：用户签到

```
用户ID: 任意 → Bitmap（userID+日期组合）
需要位运算 → Bitmap
只要计数 → 简单计数器
```

### 场景：附近搜索

```
任何位置数据 → Geospatial
数据量<100万 → 单key
数据量>100万 → 分片
```

## 八、内存计算

### HyperLogLog
```
固定 12KB，无论多少数据
```

### Bitmap
```
内存 = ceil(最大ID / 8) 字节

例如：
- 1亿用户(ID: 1-100,000,000)
- 内存 = 100,000,000 / 8 = 12.5 MB
```

### Geospatial
```
内存 ≈ 50字节 × 位置数量

例如：
- 10万骑手
- 内存 ≈ 50 × 100,000 = 5 MB
```

## 九、快速上手

### 1分钟快速体验

```bash
# 运行项目
go run .

# 选择 5 - 快速对比演示
```

### 5分钟深入了解

```bash
# 选择 6 - 三种数据类型综合对比
```

### 30分钟完整学习

```bash
# 选择 4 - 运行所有示例
```

---

**记住这三句话：**

1. HyperLogLog - "只要数量，不要人"
2. Bitmap - "连续整数布尔值"  
3. Geospatial - "位置数据专用"

**选择口诀：**

```
计数选HLL，布尔选Bitmap，
位置选Geo，列表选Set。
```
