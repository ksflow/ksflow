## Topic Names

### Kafka Multi-tenancy

Since Kafka topics [cannot be renamed](https://issues.apache.org/jira/browse/KAFKA-2333) and Kafka ACLs use topic
names for authorization decisions, how topic names are initially set matters.

Kafka provides guidance for [multi-tenancy topic naming](https://kafka.apache.org/documentation/#multitenancy-topic-naming).

### Kubernetes Multi-tenancy

Kubernetes uses namespaces to organize workloads, which is useful for authorization.

Kubernetes provides documentation on [multi-tenancy](https://kubernetes.io/docs/concepts/security/multi-tenancy/).

### Ksflow Namespace-Topic Mapping

Ksflow provides a namespaced `KafkaTopic` CRD to support kubernetes cluster multi-tenancy.

To ensure that multiple KafkaTopics don't attempt to configure the same topic, ksflow uses the KafkaTopic namespace and
name to construct the Kafka topic name. Additionally, a `ClusterKafkaConfig` specifies the topic prefix to be used.
The resulting topic in Kafka has the following structure: `<my-clusterkafkaconfig-name>.<my-kafkatopic-namespace>.<my-kafkatopic-name>`.

For example, let's consider the following topic naming structure:

    <organization>.<team>.<dataset>-<event-name>
    (e.g., "acme.infosec.telemetry-logins")

This could be accomplished with the following:

```yaml
apiVersion: ksflow.io/v1alpha1
kind: ClusterKafkaConfig
metadata:
  name: default
spec:
  topicPrefix: acme
  configs:
    bootstrap.servers: b-1.mycluster.123z8u.c2.kafka.us-east-1.amazonaws.com:9094,b-2.mycluster.123z8u.c2.kafka.us-east-1.amazonaws.com:9094
    security.protocol: SSL
---
apiVersion: ksflow.io/v1alpha1
kind: KafkaTopic
metadata:
  name: telemetry-logins
  namespace: infosec
```

When running multiple Kubernetes clusters pointed at the same Kafka cluster:
* If globally unique namespace names are enforced then each "default" ClusterKafkaConfig can use the same topicPrefix if desired.
* If namespace names are not globally unique, then do one of the following:
  1. Set a different topicPrefix for clusters that might have namespaces with the same names.
  2. Keep the same topicPrefix for clusters but avoid defining `KafkaTopics` and `KafkaACLs` in the same namespaces in
  multiple clusters with the same topicPrefix.

For specifics on how topic names are used for authorization in ksflow, see [security.md](./security.md).
