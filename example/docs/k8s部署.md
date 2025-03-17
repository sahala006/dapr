
(简体中文|[English](./README.md))


## 准备工作
### 1. 部署harbor仓库
这里自行部署，部署后默认用户名：admin，密码：Harbor12345，进入界面后创建项目`oms`




### 2. 构建服务镜像
接下来的三个服务都共用这个镜像。
```bash
cd example/oms
docker build -t 127.0.0.1:5100/devops/oms:v1 .
docker push  127.0.0.1:5100/devops/oms:v1
```

### 3. 构建dapr镜像
```bash
go build -tags "allcomponents" -o daprd cmd/daprd/main.go
docker build -t 127.0.0.1:5100/devops/dapr:v1 .
docker push 127.0.0.1:5100/devops/dapr:v1
```


### 4. 启动dapr-sidecar-injector服务

```bash
git clone https://github.com/sahala006/dapr-sidecar-injector.git
cd dapr-sidecar-injector
kubectl -f deploy/admission.yaml
go run main.go --dapr-image=172.17.0.1:5100/devops/dapr:v1 --consul-scheme=HTTP --consul-host=172.17.0.1 --consul-port=8500
```
该服务实现了对于在包含`dapr-inject=enabled`标签的namespace中创建POD时，自动注入dapr容器。具体功能请参考[https://github.com/sahala006/dapr-sidecar-injector](https://github.com/sahala006/dapr-sidecar-injector "https://github.com/sahala006/dapr-sidecar-injector/blob/master/README_zh.md")


### 5. 创建docker-registry类型的secret
```bash
kubectl create secret docker-registry harbor-token --docker-server=172.17.0.1:5100 --docker-username=admin --docker-password=Harbor12345 -n default
```


### 6. 部署服务
svc1的配置清单(example/k8s/svc1.yaml)：
```yaml
apiVersion: apps/v1
kind: Deployment

metadata:
  name: svc1-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: svc1
  template:
    metadata:
      labels:
        app: svc1
    spec:
      imagePullSecrets:
        - name: harbor-token
      containers:
        - name: svc1
          image: 172.17.0.1:5100/devops/oms:v1
```

svc2的配置清单(example/k8s/svc2.yaml)：
```yaml
apiVersion: apps/v1
kind: Deployment

metadata:
  name: svc2-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: svc2
  template:
    metadata:
      labels:
        app: svc2
    spec:
      imagePullSecrets:
        - name: harbor-token
      containers:
        - name: svc2
          image: 172.17.0.1:5100/devops/oms:v1
```

带灰度标签(blue)的配置清单(example/k8s/svc2_blue.yaml)：
```yaml
apiVersion: apps/v1
kind: Deployment

metadata:
  name: svc2-blue-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: svc2
  template:
    metadata:
      annotations:
	    service.tag: blue  #这里标注灰度标签
      labels:
        app: svc2
    spec:
      imagePullSecrets:
        - name: harbor-token
      containers:
        - name: svc2
          image: 172.17.0.1:5100/devops/oms:v1
```

#### 7. 创建k8s service用于测试
```yaml
kind: Service
apiVersion: v1
metadata:
  name: oms-service
spec:
  selector:
    app: svc1
  clusterIP: "10.96.96.96"
  type: NodePort
  ports:
  - name: web-port
    port: 8100
    targetPort: 8100 
    nodePort: 32011

  - name: dapr-http-port
    port: 3500
    targetPort: 3500
    nodePort: 32012
```


### 测试
```bash
curl -H "username:sahala" -d "age=18" http://${node_ip}:32012/v1.0/invoke/dapr_demo_svc2/method/home?a=1

```