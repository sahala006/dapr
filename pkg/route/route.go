package route

import (
	"encoding/json"
	"fmt"
	invokev1 "github.com/dapr/dapr/pkg/messaging/v1"
	consul_api "github.com/hashicorp/consul/api"
	"io"
	"net/url"
	"sync"
)

var Routes *SafeRouteMap

type VersionInfo struct {
	Old string `json:"old"`
	New string `json:"new"`
}

type Rule struct {
	Condition string `json:"condition"`
	RestItems []Item `json:"rest_items"`
}

type Item struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type RoutePolicy struct {
	Rate     int    `json:"rate"`
	RuleList []Rule `json:"rule_list"`
}

type Route struct {
	VersionInfo VersionInfo `json:"version_info"`
	RoutePolicy RoutePolicy `json:"route_policy"`
}

type SafeRouteMap struct {
	sync.RWMutex
	R map[string]*Route
}

func (this *SafeRouteMap) Update(kvs consul_api.KVPairs) {
	this.Lock()
	defer this.Unlock()
	m := make(map[string]*Route)
	for _, item := range kvs {
		var route Route
		err := json.Unmarshal(item.Value, &route)
		if err == nil {
			m[item.Key] = &route
		}
	}
	this.R = m
}

func (this *SafeRouteMap) Get(key string) *Route {
	this.RLock()
	defer this.RUnlock()
	return this.R[key]
}

func (this *SafeRouteMap) NewRoute() {

}

func (this *SafeRouteMap) SelectEntries(name string, sns []*consul_api.ServiceEntry, req *invokev1.InvokeMethodRequest) ([]*consul_api.ServiceEntry, error) {
	var versionInfo VersionInfo
	key := fmt.Sprintf("route/%s", name)
	route := this.Get(key)
	if route == nil {
		route = &Route{
			VersionInfo: VersionInfo{
				Old: "_base",
				New: "blue",
			},
		}
	}
	versionInfo = route.VersionInfo
	var snMap = make(map[string][]*consul_api.ServiceEntry)
	snMap["blue"] = make([]*consul_api.ServiceEntry, 0)
	snMap["_base"] = make([]*consul_api.ServiceEntry, 0)
	for _, sn := range sns {
		if contains(sn.Service.Tags, "blue") {
			snMap["blue"] = append(snMap["blue"], sn)
		} else {
			snMap["_base"] = append(snMap["_base"], sn)
		}
	}
	if len(route.RoutePolicy.RuleList) == 0 {
		// 按比例准发
		lb := NewWeightedLoadBalancer()
		lb.Add(versionInfo.Old, 100-route.RoutePolicy.Rate)
		lb.Add(versionInfo.New, route.RoutePolicy.Rate)
		tag := lb.Choose()
		sns, _ := snMap[tag]
		return sns, nil
	} else {
		// 按条件转发
		b, _ := io.ReadAll(req.RawData())
		body := string(b)
		metadata := req.Metadata()
		data := make(map[string]string)
		if req.ContentType() == "application/json" {
			_ = json.Unmarshal(b, &data)
		} else if req.ContentType() == "application/x-www-form-urlencoded" {
			values, _ := url.ParseQuery(body)
			for k, v := range values {
				data[k] = v[0]
			}
		}
		headers := make(map[string]string)
		for k, v := range metadata {
			headers[k] = v.Values[0]
		}

		conditionList := make([]bool, len(route.RoutePolicy.RuleList))
		for index, rule := range route.RoutePolicy.RuleList {
			conditionList[index] = judge(rule, data, headers)
		}
		if checkConditionList(conditionList, "AND") {
			return snMap[versionInfo.New], nil
		} else {
			return snMap[versionInfo.Old], nil
		}
	}
}

func judge(rule Rule, data, headers map[string]string) bool {
	conditionList := make([]bool, len(rule.RestItems))
	for index, item := range rule.RestItems {
		conditionList[index] = judgeItem(item, data, headers)
	}
	return checkConditionList(conditionList, rule.Condition)
}

func checkConditionList(conditionList []bool, match string) bool {
	if match == "AND" {
		result := true
		for _, condition := range conditionList {
			result = result && condition
		}
		return result
	} else {
		result := false
		for _, condition := range conditionList {
			result = result || condition
		}
		return result
	}
}

func judgeItem(item Item, data, headers map[string]string) bool {
	var value string
	if item.Type == "param" {
		value = data[item.Name]
	} else {
		value = headers[item.Name]
	}
	if item.Operator == "==" {
		return value == item.Value
	} else {
		return value != item.Value
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func Init(consulAddr string) {
	Routes = &SafeRouteMap{
		R: make(map[string]*Route),
	}
	go startWatcher(consulAddr)
}
