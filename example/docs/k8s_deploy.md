
(English|[简体中文](./k8s部署.md))


## Preparation
### 1. Deploy harbor repository
Deploy by yourself, default username: `admin`， password: `Harbor12345`， Create a project 'oms' after entering the page`



### 2. Build service docker image
The next three services will share this image

```bash
cd example/oms
docker build -t 127.0.0.1:5100/devops/oms:v1 .
docker push  127.0.0.1:5100/devops/oms:v1
```

### 3. Build dapr docker image
```bash
go build -tags "allcomponents" -o daprd cmd/daprd/main.go
docker build -t 127.0.0.1:5100/devops/dapr:v1 .
docker push 127.0.0.1:5100/devops/dapr:v1
```


### 4. Startup dapr-sidecar-injector service

```bash
git clone https://github.com/sahala006/dapr-sidecar-injector.git
cd dapr-sidecar-injector
kubectl -f deploy/admission.yaml
go run main.go --dapr-image=172.17.0.1:5100/devops/dapr:v1 --consul-scheme=HTTP --consul-host=172.17.0.1 --consul-port=8500
```
This service implements automatic injection of dapr containers when creating a POD in a namespace containing the `dapr-inject=enabled` label. For specific functions, please refer to[https://github.com/sahala006/dapr-sidecar-injector](https://github.com/sahala006/dapr-sidecar-injector "https://github.com/sahala006/dapr-sidecar-injector")


### 5. Create a docker-registry type secret
```bash
kubectl create secret docker-registry harbor-token --docker-server=172.17.0.1:5100 --docker-username=admin --docker-password=Harbor12345 -n default
```


### 6. Deploy service
manifest of svc1(example/k8s/svc1.yaml)：
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

manifest of svc2(example/k8s/svc2.yaml)：
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

manifest of svc2 with grayscale label(blue)(example/k8s/svc2_blue.yaml)：
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
	    service.tag: blue  # Label grayscale here
      labels:
        app: svc2
    spec:
      imagePullSecrets:
        - name: harbor-token
      containers:
        - name: svc2
          image: 172.17.0.1:5100/devops/oms:v1
```

#### 7. Create k8s service for test
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


### Test
```bash
curl -H "username:sahala" -d "age=18" http://${node_ip}:32012/v1.0/invoke/dapr_demo_svc2/method/home?a=1

```