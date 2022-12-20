## Quick Start

#### Prerequisites
* Kubernetes (i.e. [k3d](https://k3d.io/v5.4.6/#installation), [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation), etc.)
* [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)

#### Install
```shell
# install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.1/cert-manager.yaml

# install self-signed certs, kafka, ksflow
kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-install.yaml
```

#### Create a Topic
```shell
# create a KafkaTopic
kubectl apply -f - <<EOF
apiVersion: ksflow.io/v1alpha1
kind: KafkaTopic
metadata:
  name: quickstart
spec: {}
EOF

# the status should show as "Available", indicating the topic was created successfully
kubectl get kt

# the topics can be listed directly from kafka
kubectl exec -n ksflow-quickstart deploy/kafka -- /bin/sh -c "/opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server kafka.ksflow-quickstart.svc.cluster.local:9092 --list --command-config /opt/bitnami/kafka/config/admin.properties"

# details about the topic will show up in the status
kubectl get kt quickstart -oyaml
```

#### Delete the Topic
```shell
# delete the KafkaTopic
kubectl delete kt quickstart

# verify the topic was deleted from kafka
kubectl exec -n ksflow-quickstart deploy/kafka -- /bin/sh -c "/opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server kafka.ksflow-quickstart.svc.cluster.local:9092 --list --command-config /opt/bitnami/kafka/config/admin.properties"
```

#### Uninstall
```shell
# uninstall ksflow
kubectl delete -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-install.yaml

# uninstall cert-manager
kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.1/cert-manager.yaml
```
