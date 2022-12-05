# Quick Start

**The quickstart is not a production-ready installation. Among other things, it installs a kafka cluster on kubernetes
without persistence to ensure it can run on clusters without storage. DO NOT use this installation as-is for processing
any important data. See [install](install.md) for a production install.**

Before getting started you will need the following binaries on your path:
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) with access to a Kubernetes cluster.
- [helm](https://github.com/helm/helm)
- [helmfile](https://github.com/helmfile/helmfile)

 **ephemeral** schema-registry and kafka clusters, **DO NOT store any important data on these**.
You will want to replace these with actual .

Below is an example of installing all the necessary components on a new Kubernetes cluster.
Your installation may vary depending on your setup and requirements.

## Install
Create a file named `helmfile.yaml` with the following contents:
```yaml
repositories:
- name: kedacore
  url: https://kedacore.github.io/charts
- name: ksflow
  url: https://ksflow.github.io/ksflow-helm
releases:
- name: keda
  namespace: keda
  createNamespace: true
  chart: kedacore/keda
  version: 2.8.2
- name: ksflow
  namespace: ksflow
  createNamespace: true
  chart: ksflow/ksflow
  version: 0.1.0
```
To install:
```shell
helmfile sync
```
To uninstall:
```shell
helmfile destroy
kubectl delete ns keda strimzi ksflow
```
