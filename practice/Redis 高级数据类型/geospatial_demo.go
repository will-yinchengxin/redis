package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Location 表示一个地理位置
type Location struct {
	Name      string
	Longitude float64
	Latitude  float64
}

// GeospatialDemo 演示 Geospatial 的各种使用场景
type GeospatialDemo struct {
	client *redis.Client
	ctx    context.Context
}

// NewGeospatialDemo 创建 Geospatial 演示实例
func NewGeospatialDemo(client *redis.Client) *GeospatialDemo {
	return &GeospatialDemo{
		client: client,
		ctx:    context.Background(),
	}
}

// Example1_BasicUsage 基础使用示例
func (g *GeospatialDemo) Example1_BasicUsage() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Geospatial 示例 1: 基础使用")
	fmt.Println(strings.Repeat("=", 60))

	key := "geo:demo:basic"
	g.client.Del(g.ctx, key)

	// 添加一些知名地点
	locations := []Location{
		{"天安门", 116.397428, 39.909186},
		{"故宫", 116.403119, 39.918058},
		{"天坛", 116.407526, 39.882217},
		{"颐和园", 116.275199, 39.992313},
		{"鸟巢", 116.402984, 39.992831},
	}

	// 批量添加位置
	for _, loc := range locations {
		err := g.client.GeoAdd(g.ctx, key, &redis.GeoLocation{
			Name:      loc.Name,
			Longitude: loc.Longitude,
			Latitude:  loc.Latitude,
		}).Err()
		if err != nil {
			fmt.Printf("添加位置失败: %v\n", err)
			return
		}
	}

	fmt.Printf("成功添加 %d 个地点\n\n", len(locations))

	// 获取位置的经纬度
	fmt.Println("【获取位置坐标】")
	pos, err := g.client.GeoPos(g.ctx, key, "故宫", "颐和园").Result()
	if err != nil {
		fmt.Printf("获取位置失败: %v\n", err)
		return
	}
	for i, p := range pos {
		name := []string{"故宫", "颐和园"}[i]
		if p != nil {
			fmt.Printf("%s: 经度 %.6f, 纬度 %.6f\n", name, p.Longitude, p.Latitude)
		}
	}

	// 计算两点之间的距离
	fmt.Println("\n【计算距离】")
	dist, err := g.client.GeoDist(g.ctx, key, "天安门", "故宫", "km").Result()
	if err != nil {
		fmt.Printf("计算距离失败: %v\n", err)
		return
	}
	fmt.Printf("天安门到故宫的距离: %.2f 公里\n", dist)

	dist2, _ := g.client.GeoDist(g.ctx, key, "天安门", "颐和园", "km").Result()
	fmt.Printf("天安门到颐和园的距离: %.2f 公里\n", dist2)

	fmt.Println("\n✓ 基础操作演示完成")
}

// Example2_FindNearby 查找附近位置示例
func (g *GeospatialDemo) Example2_FindNearby() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Geospatial 示例 2: 查找附近的位置")
	fmt.Println(strings.Repeat("=", 60))

	key := "geo:demo:nearby"
	g.client.Del(g.ctx, key)

	// 添加北京的一些餐厅（模拟数据）
	restaurants := []Location{
		{"全聚德烤鸭店", 116.407526, 39.904989},
		{"东来顺饭庄", 116.404269, 39.906805},
		{"海底捞火锅", 116.410982, 39.908154},
		{"西贝莜面村", 116.398744, 39.915309},
		{"外婆家", 116.419863, 39.903012},
		{"绿茶餐厅", 116.395234, 39.895742},
		{"新辣道鱼火锅", 116.413452, 39.919234},
		{"小吊梨汤", 116.402345, 39.912456},
	}

	for _, loc := range restaurants {
		g.client.GeoAdd(g.ctx, key, &redis.GeoLocation{
			Name:      loc.Name,
			Longitude: loc.Longitude,
			Latitude:  loc.Latitude,
		})
	}

	fmt.Printf("添加了 %d 家餐厅\n", len(restaurants))

	// 场景：用户在天安门位置，查找方圆 2 公里内的餐厅
	userLng := 116.397428
	userLat := 39.909186
	radius := 2.0

	fmt.Printf("\n用户当前位置: 经度 %.6f, 纬度 %.6f (天安门)\n", userLng, userLat)
	fmt.Printf("搜索半径: %.1f 公里\n\n", radius)

	// 使用 GEORADIUS 查找
	results, err := g.client.GeoRadius(g.ctx, key, userLng, userLat, &redis.GeoRadiusQuery{
		Radius:      radius,
		Unit:        "km",
		WithCoord:   true,  // 返回坐标
		WithDist:    true,  // 返回距离
		Count:       10,    // 最多返回 10 个
		Sort:        "ASC", // 按距离从近到远排序
	}).Result()

	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	fmt.Printf("找到 %d 家附近的餐厅:\n", len(results))
	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Name)
		fmt.Printf("   距离: %.2f 公里\n", result.Dist)
		fmt.Printf("   坐标: (%.6f, %.6f)\n", result.Longitude, result.Latitude)
	}

	fmt.Println("\n✓ 这就是外卖 App、点评 App 查找附近商家的核心功能")
}

// Example3_FindByMember 根据成员查找附近示例
func (g *GeospatialDemo) Example3_FindByMember() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Geospatial 示例 3: 根据已知位置查找附近")
	fmt.Println(strings.Repeat("=", 60))

	key := "geo:demo:tourist"
	g.client.Del(g.ctx, key)

	// 添加北京景点
	attractions := []Location{
		{"故宫", 116.403119, 39.918058},
		{"天坛", 116.407526, 39.882217},
		{"颐和园", 116.275199, 39.992313},
		{"圆明园", 116.302763, 40.008073},
		{"香山公园", 116.189488, 39.991375},
		{"北海公园", 116.388705, 39.928349},
		{"景山公园", 116.398055, 39.926642},
		{"什刹海", 116.383331, 39.936904},
	}

	for _, loc := range attractions {
		g.client.GeoAdd(g.ctx, key, &redis.GeoLocation{
			Name:      loc.Name,
			Longitude: loc.Longitude,
			Latitude:  loc.Latitude,
		})
	}

	// 用户在"故宫"，查找周围 5 公里内的其他景点
	fmt.Println("用户当前在: 故宫")
	fmt.Println("查找周围 5 公里内的其他景点:\n")

	results, err := g.client.GeoRadiusByMember(g.ctx, key, "故宫", &redis.GeoRadiusQuery{
		Radius:   5,
		Unit:     "km",
		WithDist: true,
		Sort:     "ASC",
	}).Result()

	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	for i, result := range results {
		if result.Name != "故宫" { // 排除自己
			fmt.Printf("%d. %s - 距离 %.2f 公里\n", i, result.Name, result.Dist)
		}
	}

	fmt.Println("\n✓ 这个功能适合旅游 App 推荐附近景点")
}

// Example4_RideHailing 网约车/外卖配送场景
func (g *GeospatialDemo) Example4_RideHailing() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Geospatial 示例 4: 网约车/外卖配送场景")
	fmt.Println(strings.Repeat("=", 60))

	key := "geo:riders:beijing"
	g.client.Del(g.ctx, key)

	// 模拟 50 个骑手在北京市中心的位置
	fmt.Println("初始化 50 个骑手位置...")
	baseLocation := Location{
		Longitude: 116.397428, // 天安门经度
		Latitude:  39.909186,  // 天安门纬度
	}

	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= 50; i++ {
		// 在基准位置周围随机偏移（约 ±5 公里范围）
		rider := Location{
			Name:      fmt.Sprintf("骑手%03d", i),
			Longitude: baseLocation.Longitude + (rand.Float64()-0.5)*0.1,
			Latitude:  baseLocation.Latitude + (rand.Float64()-0.5)*0.1,
		}

		g.client.GeoAdd(g.ctx, key, &redis.GeoLocation{
			Name:      rider.Name,
			Longitude: rider.Longitude,
			Latitude:  rider.Latitude,
		})
	}

	fmt.Println("骑手位置初始化完成\n")

	// 场景：用户下单，需要找最近的 3 个骑手进行派单
	orderLocation := Location{
		Name:      "用户订单位置",
		Longitude: 116.405285,
		Latitude:  39.904989,
	}

	fmt.Printf("📦 新订单位置: 经度 %.6f, 纬度 %.6f\n", orderLocation.Longitude, orderLocation.Latitude)
	fmt.Println("正在查找最近的 3 个骑手...\n")

	// 查找 3 公里内的最近 3 个骑手
	nearbyRiders, err := g.client.GeoRadius(g.ctx, key,
		orderLocation.Longitude, orderLocation.Latitude,
		&redis.GeoRadiusQuery{
			Radius:   3,
			Unit:     "km",
			WithDist: true,
			Count:    3,
			Sort:     "ASC",
		}).Result()

	if err != nil {
		fmt.Printf("查找骑手失败: %v\n", err)
		return
	}

	if len(nearbyRiders) == 0 {
		fmt.Println("❌ 附近没有可用骑手")
		return
	}

	fmt.Println("✅ 找到以下骑手（按距离排序）:")
	for i, rider := range nearbyRiders {
		fmt.Printf("%d. %s - 距离 %.2f 公里 - 预计 %d 分钟到达\n",
			i+1, rider.Name, rider.Dist, int(rider.Dist*3)) // 假设骑手速度 20km/h
	}

	// 模拟选择最近的骑手接单
	selectedRider := nearbyRiders[0].Name
	fmt.Printf("\n🚀 系统自动派单给最近的骑手: %s\n", selectedRider)

	fmt.Println("\n✓ 这就是外卖、网约车 App 的核心派单逻辑")
}

// Example5_DynamicUpdate 动态更新位置示例
func (g *GeospatialDemo) Example5_DynamicUpdate() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Geospatial 示例 5: 动态更新位置（模拟实时定位）")
	fmt.Println(strings.Repeat("=", 60))

	key := "geo:realtime:driver"
	g.client.Del(g.ctx, key)

	// 司机初始位置
	driverName := "司机A"
	startLocation := Location{
		Name:      driverName,
		Longitude: 116.397428,
		Latitude:  39.909186,
	}

	fmt.Printf("司机初始位置: (%.6f, %.6f)\n", startLocation.Longitude, startLocation.Latitude)

	// 添加初始位置
	g.client.GeoAdd(g.ctx, key, &redis.GeoLocation{
		Name:      driverName,
		Longitude: startLocation.Longitude,
		Latitude:  startLocation.Latitude,
	})

	// 模拟司机移动 5 次
	fmt.Println("\n模拟司机移动轨迹:")
	for i := 1; i <= 5; i++ {
		time.Sleep(500 * time.Millisecond) // 模拟时间流逝

		// 每次向东北方向移动一点
		newLng := startLocation.Longitude + float64(i)*0.005
		newLat := startLocation.Latitude + float64(i)*0.005

		// 更新位置（使用 GeoAdd 覆盖旧位置）
		g.client.GeoAdd(g.ctx, key, &redis.GeoLocation{
			Name:      driverName,
			Longitude: newLng,
			Latitude:  newLat,
		})
		
		fmt.Printf("第 %d 次更新: (%.6f, %.6f)\n", i, newLng, newLat)
	}

	// 获取最终位置
	finalPos, _ := g.client.GeoPos(g.ctx, key, driverName).Result()
	if len(finalPos) > 0 && finalPos[0] != nil {
		fmt.Printf("\n司机最终位置: (%.6f, %.6f)\n", 
			finalPos[0].Longitude, finalPos[0].Latitude)
	}

	fmt.Println("\n✓ 实际应用中，司机/骑手每 5-10 秒上报一次位置")
	fmt.Println("✓ Redis Geospatial 可以实时更新，用户端可以看到实时位置")
}

// Example6_AreaQuery 区域查询示例
func (g *GeospatialDemo) Example6_AreaQuery() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Geospatial 示例 6: 区域查询（矩形范围）")
	fmt.Println(strings.Repeat("=", 60))

	key := "geo:demo:shops"
	g.client.Del(g.ctx, key)

	// 添加多个商铺
	shops := []Location{
		{"商铺A", 116.397428, 39.909186},
		{"商铺B", 116.405285, 39.904989},
		{"商铺C", 116.410982, 39.908154},
		{"商铺D", 116.395234, 39.915309},
		{"商铺E", 116.388705, 39.928349},
	}

	for _, shop := range shops {
		g.client.GeoAdd(g.ctx, key, &redis.GeoLocation{
			Name:      shop.Name,
			Longitude: shop.Longitude,
			Latitude:  shop.Latitude,
		})
	}

	fmt.Printf("添加了 %d 个商铺\n\n", len(shops))

	// 使用 GEOSEARCH 进行矩形范围查询（Redis 6.2+）
	fmt.Println("尝试使用 GEOSEARCH 进行矩形范围查询...")
	fmt.Println("（如果 Redis 版本 < 6.2，此功能不可用）")

	searchQuery := redis.GeoSearchQuery{
		Longitude:  116.405285,
		Latitude:   39.904989,
		Radius:     2,
		RadiusUnit: "km",
		Sort:       "ASC",
	}

	results, err := g.client.GeoSearch(g.ctx, key, &searchQuery).Result()
	if err != nil {
		fmt.Printf("⚠️  查询失败（可能是 Redis 版本过低）: %v\n", err)
	} else {
		fmt.Printf("找到 %d 个商铺:\n", len(results))
		for _, result := range results {
			fmt.Printf("- %s\n", result)
		}
	}

	fmt.Println("\n✓ GEOSEARCH 是更强大的搜索命令，支持矩形范围查询")
}

// Example7_GeoHash 获取 GeoHash 编码
func (g *GeospatialDemo) Example7_GeoHash() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Geospatial 示例 7: GeoHash 编码")
	fmt.Println(strings.Repeat("=", 60))

	key := "geo:demo:geohash"
	g.client.Del(g.ctx, key)

	// 添加几个著名地点
	locations := []Location{
		{"天安门", 116.397428, 39.909186},
		{"故宫", 116.403119, 39.918058},
		{"天坛", 116.407526, 39.882217},
	}

	for _, loc := range locations {
		g.client.GeoAdd(g.ctx, key, &redis.GeoLocation{
			Name:      loc.Name,
			Longitude: loc.Longitude,
			Latitude:  loc.Latitude,
		})
	}

	// 获取 GeoHash 编码
	fmt.Println("地点的 GeoHash 编码:")
	hashes, err := g.client.GeoHash(g.ctx, key, "天安门", "故宫", "天坛").Result()
	if err != nil {
		fmt.Printf("获取 GeoHash 失败: %v\n", err)
		return
	}

	names := []string{"天安门", "故宫", "天坛"}
	for i, hash := range hashes {
		fmt.Printf("%s: %s\n", names[i], hash)
	}

	fmt.Println("\n✓ GeoHash 是一种空间索引编码方式")
	fmt.Println("✓ 相近的位置有相似的 GeoHash 前缀")
	fmt.Println("✓ 可以看到天安门和故宫的 GeoHash 前缀很相似（wx4g0）")
}

// Example8_RemoveLocation 删除位置
func (g *GeospatialDemo) Example8_RemoveLocation() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Geospatial 示例 8: 删除位置")
	fmt.Println(strings.Repeat("=", 60))

	key := "geo:demo:remove"
	g.client.Del(g.ctx, key)

	// 添加几个位置
	g.client.GeoAdd(g.ctx, key,
		&redis.GeoLocation{Name: "位置A", Longitude: 116.397428, Latitude: 39.909186},
		&redis.GeoLocation{Name: "位置B", Longitude: 116.405285, Latitude: 39.904989},
		&redis.GeoLocation{Name: "位置C", Longitude: 116.410982, Latitude: 39.908154},
	)

	// 查看所有位置（使用底层 ZSet 命令）
	fmt.Println("初始位置列表:")
	members, _ := g.client.ZRange(g.ctx, key, 0, -1).Result()
	for _, member := range members {
		fmt.Printf("- %s\n", member)
	}

	// 删除位置（Geospatial 底层是 ZSet，所以用 ZREM）
	fmt.Println("\n删除 '位置B'...")
	g.client.ZRem(g.ctx, key, "位置B")

	// 再次查看
	fmt.Println("\n删除后的位置列表:")
	members, _ = g.client.ZRange(g.ctx, key, 0, -1).Result()
	for _, member := range members {
		fmt.Printf("- %s\n", member)
	}

	fmt.Println("\n✓ Geospatial 底层使用 Sorted Set（ZSet）存储")
	fmt.Println("✓ 可以使用 ZREM 删除位置，ZCARD 查看数量等")
}

// RunAllExamples 运行所有示例
func (g *GeospatialDemo) RunAllExamples() {
	fmt.Println("\n🚀 开始运行 Geospatial 所有示例...")

	g.Example1_BasicUsage()
	g.Example2_FindNearby()
	g.Example3_FindByMember()
	g.Example4_RideHailing()
	g.Example5_DynamicUpdate()
	g.Example6_AreaQuery()
	g.Example7_GeoHash()
	g.Example8_RemoveLocation()

	fmt.Println("\n✅ Geospatial 所有示例运行完成！")
}
