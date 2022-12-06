![Ksflow](images/ksflow-logo-3800x670-transparent.png)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Ksflow is a tool to simplify configuration management related to running Kubernetes pods that process Kafka topics.

Concerns such as security, configuration and scaling are standardized for developer's pods in a language-agnostic way.

## Features
#### Simple
- No Kafka proxies
- No SDKs
- Don't replace Kafka client libraries
#### Secure
- Kafka ACLs
- mTLS managed by [cert-manager](https://github.com/cert-manager/cert-manager)
- No additional developer software dependencies
- cert-manager
#### Cheap
- No required sidecars or PVs for pods
- Single container written in Go
#### Scale
- Autoscaling managed by [KEDA](https://github.com/kedacore/keda)
- Multi-cluster supported by leveraging topic prefixes
#### Developer Friendly
- Agnostic to programming language, kafka-client, and cloud-provider/on-prem
- Kubernetes-native (CRDs)
- Schema registry support

## Documentation
- [CRDs](./docs/crds.md)
- [Dependencies](./docs/dependencies.md)
- Install with the [ksflow helm chart](https://github.com/ksflow/ksflow-helm/charts/ksflow)
