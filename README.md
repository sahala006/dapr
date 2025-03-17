
(English|[简体中文](./README_zh.md))


## Function Introduction
Based on giitHub/dapp/dapp's `955436f45`, secondary development has been carried out to achieve proportional and conditional forwarding of traffic

## Preparation


### 1. Deploy Services
Deploy three services, one `SVC1` service and two `SVC2` services, and test them directly on the local machine. If you want to use the k8s deployment method, please refer to [k8s deploy](example/docs/k8s_deploy.md)。
environmental requirements： python3.7

```bash
cd example/oms
pip install -r requirements.txt
python manage.py runserver 0.0.0.0:8100  # svc1
python manage.py runserver 0.0.0.0:8200  # svc2
python manage.py runserver 0.0.0.0:8300  # svc2
```

### 2. Deploy consul
```bash
docker pull consul:1.13.1
docker run -d -p 8500:8500 --restart=always --name=consul consul:1.13.1 agent -server -bootstrap -ui -node=1 -client='0.0.0.0' 
```
Consul is used as a service registry and KV storage

### 3. Deploy dapr
```bash
# svc1 dapr：
go run  -tags=allcomponents cmd/daprd/main.go --app-id svc1 --app-port 8100 --app-protocol=http --dapr-http-port 3500 --dapr-grpc-port 3502 --metrics-port=9091 --config=example/config/.dapr/config.yaml

# svc2 dapr:
go run  -tags=allcomponents cmd/daprd/main.go --app-id svc2 --app-port 8200 --app-protocol=http --dapr-http-port 3600 --dapr-grpc-port 3602 --metrics-port=9092 --config=example/config/.dapr/config.yaml 

# svc2 dapr (with grayscale label)
go run  -tags=allcomponents cmd/daprd/main.go --app-id svc2 --app-port 8300 --app-protocol=http --dapr-http-port 3700 --dapr-grpc-port 3702 --metrics-port=9093 --config=example/config/.dapr/config_blue.yaml 
```

### Test
#### 1. access the dapr-http-port of svc1：

```bash
curl -H "username:sahala" --data "age=18" http://127.0.0.1:3500/v1.0/invoke/svc2/method/home?a=1
```
At this time, the traffic will be forwarded to two `SVC2` services in a 1:1 ratio

#### 2. Forward according to conditions
Update traffic rule：
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
        print("update key ssuccess")
    else:
        print(f"update failure:{rsp.status_code}")
if __name__ == '__main__':
    update()
```

At this point, only requests containing `env=prod` in the `--data` will be forwarded to the `SVC2` service with grayscale labels, otherwise they will be forwarded to the `SVC2` service without labels

#### 3. Forward proportionally
Update traffic rule：
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
        print("update key ssuccess")
    else:
        print(f"update failure:{rsp.status_code}")
if __name__ == '__main__':
    update()
```
At this point, 30% of the traffic will be forwarded to the `SVC2` service with grayscale labels, and 70% of the traffic will be forwarded to the `SVC2` service without labels.
