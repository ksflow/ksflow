## Dependencies

| Tool                                                   | Required | Purpose       |
|--------------------------------------------------------|----------|---------------|
| Kafka                                                  | yes      | Storage       |
| *Schema Registry                                       | yes      | Schemas       |
| [Istio](https://github.com/istio/istio)                | yes      | Security      |
| [KEDA](https://github.com/kedacore/keda)               | yes      | Autoscaling   |
| [Prometheus](https://github.com/prometheus/prometheus) | no       | Monitoring    |
| [Grafana](https://github.com/grafana/grafana)          | no       | Observability |
| [Kyverno](https://github.com/kyverno/kyverno)          | no       | Quotas        |
**Schema Registry must be compatible with Confluent SR clients (i.e. [Confluent](https://github.com/confluentinc/schema-registry), [Apicurio](https://github.com/Apicurio/apicurio-registry), [Karapace](https://github.com/aiven/karapace))*
