## Quick Start

#### Prerequisites
* Kubernetes (i.e. [k3d](https://k3d.io/v5.4.6/#installation), [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation), etc.)
* [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
* [helm](https://helm.sh/docs/intro/install/)

#### Install
```shell
# install kafka
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install kafka bitnami/kafka \
  --set persistence.enabled=false \
  --set zookeeper.persistence.enabled=false

# install ksflow
helm repo add ksflow https://ksflow.github.io/ksflow-helm
helm install ksflow ksflow/ksflow \
  --set "kafka.bootstrapServers={kafka-0:9092}"
```

#### Create a Topic
```shell
# create a KafkaTopic
kubectl apply -f - <<EOF
apiVersion: ksflow.io/v1alpha1
kind: KafkaTopic
metadata:
  name: ksflow-quickstart
spec:
  partitions: 1
  replicationFactor: 1
EOF

# the status should show as "Available", indicating the topic was created successfully
kubectl get kt

# topics can be listed using the kafka client shell scripts
kubectl exec sts/kafka -- /opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list

# details about the topic will show up in the status
kubectl get kt quickstart -oyaml
```

#### Delete the Topic
```shell
# delete the KafkaTopic
kubectl delete kt ksflow-quickstart

# verify the topic was deleted from kafka
kubectl exec sts/kafka -- /opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
```

#### Uninstall
```shell
# uninstall ksflow
helm uninstall ksflow

# uninstall kafka
helm uninstall kafka
```
