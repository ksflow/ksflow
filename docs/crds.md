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

### KafkaTopic

### KafkaACL

### KafkaConsumerConfig

### KafkaAdminClientConfig
