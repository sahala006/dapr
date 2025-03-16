package route

import (
	"math/rand"
	"time"
)

type WeightedLoadBalancer struct {
	weights map[string]int // 元素名到权重的映射
	total   int            // 总权重
}

func NewWeightedLoadBalancer() *WeightedLoadBalancer {
	return &WeightedLoadBalancer{
		weights: make(map[string]int),
		total:   0,
	}
}

func (lb *WeightedLoadBalancer) Add(name string, weight int) {
	lb.weights[name] = weight
	lb.total += weight
}

func (lb *WeightedLoadBalancer) Choose() string {
	rand.Seed(time.Now().UnixNano())    // 确保每次运行都不一样，实际应用中通常只在初始化时设置一次种子。
	chosenWeight := rand.Intn(lb.total) // 生成一个在[0, totalWeight)范围内的随机数
	currentWeight := 0
	for name, weight := range lb.weights {
		currentWeight += weight
		if chosenWeight < currentWeight {
			return name
		}
	}
	return "" // 理论上不应该执行到这里，除非没有元素或者权重总和为零。
}
