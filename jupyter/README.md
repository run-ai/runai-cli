##  Use arena in jupyter notebook in the Kubernetes cluster

1. Build jupyter notebook with `arena` by yourself

```
mkdir -p $(go env GOPATH)/src/github.com/kubeflow
cd $(go env GOPATH)/src/github.com/kubeflow
git clone https://github.com/run-ai/runai-cli.git
cd arena
make notebook-image-cpu
```

> You can update the docker repo name from `cheyang` to your docker repo

2. Deploy in Kubernetes

```
kubectl create -f https://raw.githubusercontent.com/kubeflow/arena/master/jupyter/arena-jupyter.yaml
```

> The default password is `passw0rd`, you can update it before `kubectl create`

> The RBAC Settings of the service account `arena-notebook`
