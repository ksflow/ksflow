## Quick Start

#### Prerequisites
* Kubernetes (i.e. [k3d](https://k3d.io/v5.4.6/#installation), [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation), etc.)
* [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
* [helm](https://helm.sh/docs/intro/install/)

#### Install
1. Install Ksflow using the [ksflow helm chart](https://github.com/ksflow/ksflow-helm/tree/main/charts/ksflow)
   ```bash
   helm repo add https://ksflow.github.io/ksflow-helm
   helm install --namespace ksflow-system --create-namespace ksflow ksflow/ksflow
   ```
2. Install Kafka
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-kafka.yaml
   ```
3. Configure Ksflow to know about the Kafka cluster
   ```yaml
   kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-ckcc.yaml
   ```

#### Create a Topic
TODO!!!!!!!!!!!!!!!!!!!!!!!
1. Verify the topic was created
   ```bash
   kc exec deploy/ksflow-quickstart-kafka -- /bin/sh -c "/opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list"
   ```
