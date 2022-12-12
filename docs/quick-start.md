## Quick Start

#### Prerequisites
* Kubernetes (i.e. [k3d](https://k3d.io/v5.4.6/#installation), [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation), etc.)
* [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
* [helm](https://helm.sh/docs/intro/install/)
* [kubectx](https://github.com/ahmetb/kubectx)

#### Install
```shell
# create a new namespace and switch to it
kubectl create ns ksflow-system
kubens ksflow-system

# install ksflow
helm repo add https://ksflow.github.io/ksflow-helm
helm install ksflow ksflow/ksflow

# install kafka
kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-kafka.yaml

# create a ClusterKafkaConfig
kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-ckc.yaml

# watch the ClusterKafkaConfig until it's STATUS is "Available"
kubectl get ckc --watch
```

#### Create a Topic
```shell
# create a KafkaTopic
kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-kt.yaml

# watch the KafkaTopic until it's STATUS is "Available"
kubectl get kt --watch

# verify the topic was created in kafka
kubectl exec deploy/ksflow-quickstart-kafka -- /bin/sh -c "/opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list"
```

#### Uninstall
```shell
kubens default
kubectl delete ns ksflow-system
```