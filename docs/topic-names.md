## Topic Names

### Kafka Multi-tenancy

Since Kafka topics [cannot be renamed](https://issues.apache.org/jira/browse/KAFKA-2333) and Kafka ACLs use topic
names for authorization decisions, how topic names are initially set matters.

Kafka provides guidance for [multi-tenancy topic naming](https://kafka.apache.org/documentation/#multitenancy-topic-naming).

### Kubernetes Multi-tenancy

Kubernetes uses namespaces to organize workloads, which are useful for authorization.

Kubernetes provides documentation on [multi-tenancy](https://kubernetes.io/docs/concepts/security/multi-tenancy/).

### Ksflow Namespace-Topic Mapping

Ksflow provides a namespaced `KafkaTopic` CRD to support kubernetes cluster multi-tenancy.

To ensure that multiple KafkaTopics don't attempt to configure the same topic, ksflow uses the KafkaTopic namespace and
name to construct the Kafka topic name. Additionally, a `ClusterKafkaClusterConfig` specifies the topic prefix to be used.
The resulting topic in Kafka has the following structure: `<my-clusterkafkaclusterconfig-name>.<my-kafkatopic-namespace>.<my-kafkatopic-name>`.

Let's take a look at an example provided in the Kafka documentation:
> By project or product: Here, a team manages more than one project. Their credentials will be different for each project, so all the controls and settings will always be project related.
>
> Example topic naming structure:
>
> &lt;project&gt;.&lt;product&gt;.&lt;event-name&gt;
> (e.g., "mobility.payments.suspicious")

This could be accomplished with the following:
```yaml
apiVersion: ksflow.io/v1alpha1
kind: ClusterKafkaClusterConfig
metadata:
  name: mycluster-mobility
spec:
  topicPrefix: mobility
  configs:
    bootstrapServers: b-1.mycluster.123z8u.c2.kafka.us-east-1.amazonaws.com:9094,b-2.mycluster.123z8u.c2.kafka.us-east-1.amazonaws.com:9094
  ...
---
apiVersion: ksflow.io/v1alpha1
kind: KafkaTopic
metadata:
  name: suspicious
  namespace: payments
spec:
  cluster: mycluster-mobility
  ...
```

This is just one example, configuration will depend on your use-case.

For specifics on how authorization works in ksflow, see [security.md](./security.md).
