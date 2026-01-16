# Redis 高级数据类型完全指南：HyperLogLog + Geospatial + Bitmap

## 目录
1. [HyperLogLog - 唯一计数估计](#hyperloglog)
2. [Geospatial - 地理空间索引](#geospatial)
3. [Bitmap - 位图](#bitmap)
4. [三者对比与选择](#三者对比)

---

## 一、HyperLogLog - 唯一计数估计

### 1.1 原理详解

#### 什么是 HyperLogLog？
HyperLogLog (HLL) 是一种**概率型数据结构**，用于估算集合的基数（唯一元素数量）。

#### 核心特点
- **极低内存占用**：每个 HLL 只需要 **12KB** 内存
- **误差率**：标准误差约 **0.81%**
- **适用场景**：百万、千万级别的去重计数

#### 算法原理（简化理解）

```
传统方式统计 UV（独立访客）：
- Set 存储: {user1, user2, user3, ...}
- 1亿用户 ≈ 需要 3-4GB 内存

HyperLogLog 方式：
- 使用概率算法估算
- 1亿用户 ≈ 只需 12KB 内存！
- 误差在 1% 以内
```

### 1.2 使用场景

| 场景 | 说明 | 传统方案问题 |
|------|------|--------------|
| **网站 UV 统计** | 统计每日独立访客 | 存储所有用户 ID，内存爆炸 |
| **App 日活/月活** | DAU/MAU 统计 | 需要大量内存存储用户 ID |
| **搜索关键词去重** | 统计独立搜索词数量 | 海量关键词无法存储 |
| **IP 去重统计** | 统计访问 IP 数量 | IPv6 地址空间巨大 |

### 1.3 命令速查

```redis
# 添加元素
PFADD key element [element ...]

# 获取基数估算值
PFCOUNT key [key ...]

# 合并多个 HLL
PFMERGE destkey sourcekey [sourcekey ...]
```

---

## 二、Geospatial - 地理空间索引

### 2.1 原理详解

#### 什么是 Geospatial？
Redis 的地理空间功能基于 **Sorted Set** 实现，使用 **GeoHash** 编码技术存储地理位置。

#### 核心特点
- 可以存储经纬度坐标
- 计算两点之间的距离
- 查找指定范围内的位置
- 底层使用 Sorted Set（ZSet）

### 2.2 使用场景

| 场景 | 说明 | 典型应用 |
|------|------|---------|
| **附近的人** | 查找周围用户 | 社交 App（陌陌、探探）|
| **外卖配送** | 找最近的骑手 | 美团、饿了么 |
| **打车服务** | 匹配附近司机 | 滴滴、Uber |
| **门店查询** | 搜索附近门店 | 大众点评、地图 App |

### 2.3 命令速查

```redis
# 添加地理位置
GEOADD key longitude latitude member

# 获取位置的经纬度
GEOPOS key member

# 计算两点距离
GEODIST key member1 member2 [unit]

# 范围查询
GEORADIUS key lng lat radius unit [options]
GEORADIUSBYMEMBER key member radius unit [options]
```

---

## 三、Bitmap - 位图

### 3.1 原理详解

#### 什么是 Bitmap？
Bitmap 不是一种独立的数据类型，而是基于 **String** 类型实现的位操作。它将字符串看作是由二进制位组成的数组。

#### 核心特点
- **极致内存效率**：1 个用户只占 1 bit
- **精确计数**：100% 准确
- **适用前提**：用户 ID 必须是**连续的整数**
- **位运算支持**：AND、OR、XOR、NOT

#### 工作原理

```
用户 ID → bit offset
用户 1  → bit 0
用户 2  → bit 1
用户 3  → bit 2
...

存储: 0101101... (每个 bit 表示一个用户是否活跃)
```

**内存计算公式**：
```
内存(字节) = 最大用户 ID / 8

示例：
- 1 亿用户 (ID: 1-100,000,000)
- 内存占用 = 100,000,000 / 8 = 12.5 MB
```

### 3.2 使用场景

| 场景 | 说明 | 为什么适合 Bitmap |
|------|------|------------------|
| **用户签到** | 记录每日签到状态 | 天数连续，布尔值 |
| **在线用户统计** | 统计实时在线用户 | 用户 ID 连续 |
| **活跃用户分析** | DAU 统计 | 可用位运算分析 |
| **权限管理** | 用户权限标记 | 权限位连续 |
| **A/B 测试分组** | 用户分组标记 | 布尔值标记 |

### 3.3 命令速查

```redis
# 设置位值
SETBIT key offset value

# 获取位值
GETBIT key offset

# 统计为 1 的位数量
BITCOUNT key [start end]

# 位运算
BITOP operation destkey key [key ...]
  - operation: AND, OR, XOR, NOT

# 查找第一个指定值的位置
BITPOS key bit [start end]
```

---

## 四、三者对比与选择

### 4.1 功能对比表

| 特性 | HyperLogLog | Bitmap | Geospatial |
|------|-------------|--------|-----------|
| **主要用途** | 去重计数 | 布尔值记录 | 位置存储与查询 |
| **底层实现** | 概率算法 | String（位操作） | Sorted Set + GeoHash |
| **内存占用** | 固定 12KB | ID最大值/8 | ~50字节/位置 |
| **精确度** | ~99% | 100% | 100% |
| **数据量限制** | 无限制 | 受最大 ID 影响 | 建议 < 100万/key |
| **ID 要求** | 任意 | 必须连续整数 | 无要求 |

### 4.2 内存对比（10万用户）

| 方案 | 内存占用 | 精确度 | 说明 |
|------|---------|--------|------|
| **Set** | ~3.6 MB | 100% | 可获取元素列表 |
| **Bitmap** | ~12 KB | 100% | 需要连续 ID |
| **HyperLogLog** | 12 KB | ~99% | 无法获取具体元素 |

### 4.3 选择决策树

```
┌─ 需要知道具体元素？
│  ├─ 是 → 用 Set
│  └─ 否 ┐
│        │
│        ├─ 需要精确计数？
│        │  ├─ 是 → 用户 ID 连续？
│        │  │  ├─ 是 → 用 Bitmap ✓
│        │  │  └─ 否 → 用 Set（小规模）或 HyperLogLog（大规模）
│        │  └─ 否（可接受误差）→ 用 HyperLogLog ✓
│        │
└────────┴─ 位置数据？ → 用 Geospatial ✓
```

### 4.4 典型场景推荐

#### 场景 1: 网站 UV 统计

**需求**：统计每日独立访客，用户 ID 是 UUID。

```
选择: HyperLogLog
理由: 
  ✓ 用户 ID 不连续（UUID）
  ✓ 只需要计数，不需要用户列表
  ✓ 可接受 1% 误差
  ✓ 内存占用极小
```

#### 场景 2: 用户签到系统

**需求**：记录用户每日签到状态，用户 ID 是自增整数。

```
选择: Bitmap
理由:
  ✓ 用户 ID 连续
  ✓ 只需记录布尔值（签到/未签到）
  ✓ 需要精确数据
  ✓ 内存效率高
```

#### 场景 3: 外卖骑手派单

**需求**：找到距离用户最近的骑手。

```
选择: Geospatial
理由:
  ✓ 需要位置数据
  ✓ 需要计算距离
  ✓ 需要范围查询
```

#### 场景 4: App 日活月活统计

**需求**：统计 DAU/MAU，用户 ID 是 UUID。

```
方案 A (推荐): HyperLogLog
  ✓ 只需计数
  ✓ 用户量大
  ✓ 可接受误差

方案 B: Bitmap
  条件: 如果用户 ID 可以映射到连续整数
  ✓ 需要精确值
  ✓ 可以做位运算分析（如留存率）
```

### 4.5 性能基准数据

#### HyperLogLog
```
内存: 固定 12KB
写入: ~10万次/秒
查询: O(1)
误差: 0.81%
```

#### Bitmap
```
内存: 最大ID / 8 字节
写入: ~10万次/秒
查询: O(1)
误差: 0%
```

#### Geospatial
```
内存: ~50字节/位置
写入: O(log N)
查询: O(N + log M)
建议: 单key < 100万位置
```

---

## 五、实战建议

### 5.1 组合使用

实际项目中，这三种数据类型可以组合使用：

```
电商平台完整方案:
├─ HyperLogLog: 统计每日 UV
├─ Bitmap: 记录用户签到、活跃状态
└─ Geospatial: 门店定位、配送范围查询
```

### 5.2 最佳实践

#### HyperLogLog
```go
// ✓ 按日期分 key
key := fmt.Sprintf("hll:uv:%s", date)
client.PFAdd(ctx, key, userID)
client.Expire(ctx, key, 90*24*time.Hour)

// ✓ 合并多天数据
client.PFMerge(ctx, "hll:wau", key1, key2, key3)
```

#### Bitmap
```go
// ✓ 用户签到
key := fmt.Sprintf("bitmap:signin:%d:%s", userID, month)
client.SetBit(ctx, key, int64(day-1), 1)

// ✓ 统计活跃用户
client.BitCount(ctx, key, nil)

// ✓ 计算留存（位运算）
client.BitOpAnd(ctx, destKey, key1, key2)
```

#### Geospatial
```go
// ✓ 按城市分片
key := fmt.Sprintf("geo:%s:riders", city)
client.GeoAdd(ctx, key, location)

// ✓ 限制搜索范围和数量
query := &redis.GeoRadiusQuery{
    Radius: 5,
    Unit:   "km",
    Count:  20,
}
```

### 5.3 常见错误

#### ❌ 错误 1: 用 HyperLogLog 存储小量数据
```go
// 1000 个用户用 HyperLogLog → 浪费
// 应该直接用 Set
```

#### ❌ 错误 2: 用 Bitmap 存储非连续 ID
```go
// 用户 ID: uuid → Bitmap 无法使用
// 应该用 Set 或 HyperLogLog
```

#### ❌ 错误 3: Geospatial 单 key 存储过多位置
```go
// 一个 key 存 100 万位置 → 性能下降
// 应该按城市/区域分片
```

---

## 六、总结对比

| 数据类型 | 记住这一句 | 典型口诀 |
|---------|-----------|---------|
| **HyperLogLog** | 大规模去重计数神器 | "数量不要人，误差一个点" |
| **Bitmap** | 连续 ID 的布尔值首选 | "连续整数布尔值，节省内存真精彩" |
| **Geospatial** | LBS 应用必备 | "附近的人和位置，地理数据不迷路" |

---

**选择建议总结**：

1. **只要计数** → HyperLogLog
2. **连续 ID + 布尔值** → Bitmap  
3. **位置数据** → Geospatial
4. **需要元素列表** → Set
5. **混合场景** → 组合使用
