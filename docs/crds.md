## CRDs

Ksflow introduces the following CRDs:

| CRD                         | Short Name | Namespaced | Owns          |
|-----------------------------|------------|------------|---------------|
| `ClusterKafkaClusterConfigConfig` | `ckcc`     | no         |               |
| `KafkaTopic`                | `kt`       | yes        | topic (kafka) |
| `KafkaACL`                  | `ka`       | yes        | ACL (kafka)   |
| `KafkaConsumerConfig`       | `kcc`      | yes        |               |
| `KafkaProducerConfig`       | `kpc`      | yes        |               |
| `KafkaAdminClientConfig`    | `kacc`     | yes        |               |

### ClusterKafkaClusterConfigConfig
Ksflow does not manage Kafka brokers, only the things that an application reasonably might need to configure.
Because of this, a ClusterKafkaClusterConfigConfig only defines the properties required for an application to connect to an
external Kafka cluster. Notably, it specifies:

| Name                             | Example                                   | Values                                                                                                                      |
|----------------------------------|-------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------|
| `.spec.topicPrefix`              | `com.example`                             | Labels must match [RFC 1035](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names) |
| `.spec.configs.bootstrapServers` | `k1.example.com:9092,k2.example.com:9092` |                                                                                                                             |
| `.spec.configs.securityProtocol` | `SSL`                                     | `PLAINTEXT`,`SSL`                                                                                                           |

Notes:
- **Be VERY careful if changing properties, as they will trigger updates in all objects that reference the ClusterKafkaClusterConfigConfig.**
- **Setting the securityProtocol to SSL is strongly recommended. See [security.md](./security.md) for details.**
- A kubernetes finalizer is used to prevent deletion of the ClusterKafkaClusterConfigConfig until all references are removed.

### KafkaTopic

### KafkaACL

### KafkaConsumerConfig

### KafkaAdminClientConfig
