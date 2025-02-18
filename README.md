# device-plugin-demo

使用kind快速搭建k8s集群
```bash
kind create cluster --name 1master-with-3workers --config ./deploy/kind-conf.yaml
```
构建镜像
```bash
make build-image
```

查看镜像
```bash
➜  device-plugin-demo git:(main) ✗ docker image ls | grep device-plugin-demo                           
forrest-tao/device-plugin-demo            v1                                                                            00b54ab7db2d   10 seconds ago   23.3MB
```

将构建成功的镜像加载到 kind集群中
```bash
kind load docker-image forrest-tao/device-plugin-demo:v1 --name 1master-with-3workers
```

以daemonset的形式 部署device_plugin 
```bash
k apply -f ./deploy/ds.yaml
```

查看kind 容器
```bash
➜  kind git:(main) ✗ docker ps
CONTAINER ID   IMAGE                  COMMAND                  CREATED       STATUS      PORTS                               NAMES
2baa2a65f40a   kindest/node:v1.31.0   "/usr/local/bin/entr…"   4 days ago    Up 4 days                                       1master-with-3workers-worker
b0dc68fb27d5   kindest/node:v1.31.0   "/usr/local/bin/entr…"   4 days ago    Up 4 days   127.0.0.1:56436->6443/tcp           1master-with-3workers-control-plane
9b05e588ac4a   kindest/node:v1.31.0   "/usr/local/bin/entr…"   4 days ago    Up 4 days                                       1master-with-3workers-worker3
d8179569ff66   kindest/node:v1.31.0   "/usr/local/bin/entr…"   4 days ago    Up 4 days                                       1master-with-3workers-worker2
```

进入worker1 node(容器2baa2a65f40a)，查看 `/var/lib/kubelet/device-plugins` 下的sock文件
```bash
➜  kind git:(main) ✗ docker exec -it 2baa2a65f40a sh
# cd /var/lib/kubelet/device-plugins
# ls
xpu.sock  kubelet.sock  kubelet_internal_checkpoint
```

尝试创建一个使用 xpu 自定义资源的pod
```bash
k apply -f ./deploy/test-pod.yaml
```
发现 资源pod状态，会发现 pod处于pending（因为node中demo.com/xpu 的capacity 为0）
```bash
➜  i-device-plugin git:(main) ✗ k get po
NAME      READY   STATUS    RESTARTS   AGE
xpu-pod   0/1     Pending   0          10s
➜  i-device-plugin git:(main) ✗ k describe po xpu-pod
Name:             xpu-pod
Namespace:        default
Priority:         0
Service Account:  default
Node:             <none>
Labels:           <none>
Annotations:      <none>
Status:           Pending
IP:
IPs:              <none>
Containers:
  xpu-container:
    Image:      busybox
    Port:       <none>
    Host Port:  <none>
    Command:
      sh
      -c
      echo Hello, Kubernetes! && sleep 3600
    Limits:
      demo.com/xpu:  1
    Requests:
      demo.com/xpu:  1
    Environment:     <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-8sv8r (ro)
Conditions:
  Type           Status
  PodScheduled   False
Volumes:
  kube-api-access-8sv8r:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    ConfigMapOptional:       <nil>
    DownwardAPI:             true
QoS Class:                   BestEffort
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type     Reason            Age   From               Message
  ----     ------            ----  ----               -------
  Warning  FailedScheduling  21s   default-scheduler  0/4 nodes are available: 1 node(s) had untolerated taint {node-role.kubernetes.io/control-plane: }, 3 Insufficient demo.com/xpu. preemption: 0/4 nodes are available: 1 Preemption is not helpful for scheduling, 3 No preemption victims found for incoming pod.
```

我们在node上创建一个资源
```bash
# cd /etc/xpu
# touch x1
# ls
x1
```

会发现node上资源demo.com/xpu: "1"
```bash
capacity:
    cpu: "10"
    demo.com/xpu: "1"
```

```bash
➜  device-plugin-demo git:(main) ✗ k apply -f ./deploy/test-pod.yaml
pod/xpu-pod-1 created
```

device_plugin的部分日志
```bash
I0218 09:20:42.764473       1 api.go:53] [Allocate] received request: x1
```
