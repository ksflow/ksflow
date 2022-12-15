## Quick Start

For the quickstart everything is installed in a single namespace. In practice however, this will likely not be the case.
* Ksflow should generally be installed in a dedicated namespace (i.e. `ksflow-system`, `ksflow`).
* Each `KafkaTopic` should be installed in whatever namespace is appropriate (i.e. alongside the pods producing to it).
* Kafka may or may not reside in kubernetes. The one installed in the quickstart is for demonstration purposes only.

#### Prerequisites
* Kubernetes (i.e. [k3d](https://k3d.io/v5.4.6/#installation), [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation), etc.)
* [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
* [helm](https://helm.sh/docs/intro/install/)

#### Install
```shell
# install ksflow
helm repo add https://ksflow.github.io/ksflow-helm
helm install ksflow ksflow/ksflow --set crds.keep=false

# install kafka
kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-kafka.yaml

# configure ksflow controller to use the kafka
kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-kafka-configs.yaml
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

#### Delete the Topic
```shell
# delete the KafkaTopic
kubectl delete kt quickstart

# verify the topic was deleted from kafka
kubectl exec deploy/ksflow-quickstart-kafka -- /bin/sh -c "/opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list"
```

#### Uninstall
```shell
kubectl delete -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-kafka.yaml
helm uninstall ksflow
```
