## Dependencies

| Tool                                                         | Required | Purpose       |
|--------------------------------------------------------------|----------|---------------|
| Kafka                                                        | yes      | Storage       |
| *Schema Registry                                             | yes      | Schemas       |
| [Cert Manager](https://github.com/cert-manager/cert-manager) | yes      | Security      |
| [KEDA](https://github.com/kedacore/keda)                     | yes      | Autoscaling   |
| [Prometheus](https://github.com/prometheus/prometheus)       | no       | Monitoring    |
| [Grafana](https://github.com/grafana/grafana)                | no       | Observability |
**Schema Registry must be compatible with Confluent SR clients (i.e. [Confluent](https://github.com/confluentinc/schema-registry), [Apicurio](https://github.com/Apicurio/apicurio-registry), [Karapace](https://github.com/aiven/karapace))*
