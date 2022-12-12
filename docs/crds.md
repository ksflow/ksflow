## CRDs

Ksflow introduces the following CRDs:

| CRD                      | Short Name | Namespaced | Owns          |
|--------------------------|------------|------------|---------------|
| `ClusterKafkaConfig`     | `ckc`      | no         |               |
| `KafkaTopic`             | `kt`       | yes        | topic (kafka) |
| `KafkaACL`               | `ka`       | yes        | ACL (kafka)   |
| `KafkaConsumerConfig`    | `kcc`      | yes        |               |
| `KafkaProducerConfig`    | `kpc`      | yes        |               |
| `KafkaAdminClientConfig` | `kacc`     | yes        |               |

### ClusterKafkaConfig
Ksflow does not manage Kafka brokers, only the things that an application reasonably might need to configure.
Because of this, a ClusterKafkaConfig only defines the properties required for an application to connect to an
external Kafka cluster. Notably, it specifies:

| Name                               | Example                                   | Values                                                                                                                      |
|------------------------------------|-------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------|
| `.spec.topicPrefix`                | `com.example`                             | Labels must match [RFC 1035](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names) |
| `.spec.configs.bootstrap\.servers` | `k1.example.com:9092,k2.example.com:9092` |                                                                                                                             |
| `.spec.configs.security\.protocol` | `SSL`                                     | `PLAINTEXT`,`SSL`                                                                                                           |

**Important Notes**:
- Be VERY careful if changing properties, as they will trigger updates in all objects that reference the ClusterKafkaConfig.
- Setting the "security.protocol" to SSL is strongly recommended. See [security.md](./security.md) for details.
- Take care when selecting the topicPrefix. See [topic-names.md](./topic-names.md) for details.

Currently, there is only a single cluster-scoped `ClusterKafkaConfig` named `default` that is used. This is for a couple of reasons:
1. Topic naming is based on the `KafkaTopic` name/namespace, resulting in complications if multiple KafkaConfigs exist. Such as
needing an option in the `KafkaTopic` to specify the KafkaCluster, making things less clear to the user.
2. Kafka does not provide a way to restrict ACL managers on which ACLs they can or cannot manage, so the ksflow controller
must make that decision.  Topics are created by a Job within the namespace of the `KafkaTopic`, using permissions from
a ServiceAccount in that namespace. However, due to this Kafka limitation, for ACLs that is not an option.

In the future it may make sense to support additional configuration for topics not residing within the `ClusterKafkaConfig`'s `topicPrefix`.
For example, a `ExternalKafkaTopic` which has the topicName, bootstrap.servers, and security.protocol in it's `spec`, while providing no ACL support.
This would resolve both these issues, without requiring a controller to run in every namespace.

### KafkaTopic

### KafkaACL

### KafkaConsumerConfig

### KafkaAdminClientConfig
