package route

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

func newWatcher(watchType string, opts map[string]string) (*watch.Plan, error) {
	var options = map[string]interface{}{
		"type":   watchType,
		"prefix": "route",
	}
	// 组装请求参数。(监控类型不同，其请求参数不同)
	for k, v := range opts {
		options[k] = v
	}

	wp, err := watch.Parse(options)
	if err != nil {
		return nil, err
	}

	wp.Handler = func(idx uint64, data interface{}) {

		switch d := data.(type) {
		case consulapi.KVPairs:
			fmt.Println(555555, d)
			Routes.Update(d)
		default:
			fmt.Printf("不能判断监控的数据类型: %v", &d)
		}
	}
	return wp, nil
}

func registerWatcher(watchType string, opts map[string]string, consulAddr string) error {
	wp, err := newWatcher(watchType, opts)
	defer wp.Stop()
	if err = wp.Run(consulAddr); err != nil {
		fmt.Println("err: ", err)
		return err
	}

	return nil
}

func startWatcher(consulAddr string) {
	_ = registerWatcher("keyprefix", nil, consulAddr)
}
