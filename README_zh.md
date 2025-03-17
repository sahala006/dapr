
(简体中文|[English](./README.md))

## 功能介绍
基于github.com/dapr/dapr的`955436f45`来进行二次开发，实现了流量的按比例转发和按条件转发


## 准备工作


### 1. 部署服务
部署3个服务，一个svc1服务，两个svc2服务，这里直接在本机测试，如果想用k8s的部署方式请参考[k8s部署](example/docs/k8s部署.md)。
环境要求： python3.7

```bash
cd example/oms
pip install -r requirements.txt
python manage.py runserver 0.0.0.0:8100  # 启动svc1服务
python manage.py runserver 0.0.0.0:8200  # 启动svc2服务
python manage.py runserver 0.0.0.0:8300  # 启动svc2服务
```

### 2. 部署consul
```bash
docker pull consul:1.13.1
docker run -d -p 8500:8500 --restart=always --name=consul consul:1.13.1 agent -server -bootstrap -ui -node=1 -client='0.0.0.0' 
```
consul作为服务注册中心和kv存储来使用

### 3. 部署dapr
```bash
# svc1 dapr：
go run  -tags=allcomponents cmd/daprd/main.go --app-id svc1 --app-port 8100 --app-protocol=http --dapr-http-port 3500 --dapr-grpc-port 3502 --metrics-port=9091 --config=example/config/.dapr/config.yaml

# svc2 dapr:
go run  -tags=allcomponents cmd/daprd/main.go --app-id svc2 --app-port 8200 --app-protocol=http --dapr-http-port 3600 --dapr-grpc-port 3602 --metrics-port=9092 --config=example/config/.dapr/config.yaml 

# svc2 dapr (带灰度标签)
go run  -tags=allcomponents cmd/daprd/main.go --app-id svc2 --app-port 8300 --app-protocol=http --dapr-http-port 3700 --dapr-grpc-port 3702 --metrics-port=9093 --config=example/config/.dapr/config_blue.yaml 
```

### 测试
#### 1. 访问svc1的dapr-http-port：

```bash
curl -H "username:sahala" --data "age=18" http://127.0.0.1:3500/v1.0/invoke/svc2/method/home?a=1
```
此时流量会100%转发到不带灰度标签的svc2服务。

#### 2. 按条件转发
更新流量规则：

```python
import requests
headers = {
    "contentType": "application/json"
}
def update():
    url = "http://127.0.0.1:8500/v1/kv/route/svc2"
    data = {
        "version_info": {
            "old": "_base",
            "new": "blue",
        },
        "route_policy": {
            "rate": 50,
            "rule_list": [{
                "condition": "AND",
                "rest_items": [
                    {
                        "type": "param",
                        "name": "env",
                        "operator": "==",
                        "value": "prod"
                    }
                ],
            }]
        }
    }
    rsp = requests.put(url, headers=headers, data=json.dumps(data))
    if rsp.status_code == 200:
        print("更新key成功")
    else:
        print(f"更新失败:{rsp.status_code}")
if __name__ == '__main__':
    update()
```

此时只有`--data`部分里包含`env=prod`的请求才会转发到带有灰度标签(blue)的svc2服务，否则转发到不带灰度标签的svc2服务。


#### 3. 按比例转发
更新流量规则：
```python
import requests
headers = {
    "contentType": "application/json"
}
def update():
    url = "http://127.0.0.1:8500/v1/kv/route/svc2"
    data = {
        "version_info": {
            "old": "_base",
            "new": "blue",
        },
        "route_policy": {
            "rate": 30,
        }
    }
    rsp = requests.put(url, headers=headers, data=json.dumps(data))
    if rsp.status_code == 200:
        print("更新key成功")
    else:
        print(f"更新失败:{rsp.status_code}")
if __name__ == '__main__':
    update()
```
此时30%流量会转发到带灰度标签的svc2服务，70%的流量转发到不带标签的的svc2服务。

