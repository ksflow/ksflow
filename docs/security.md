## Certificates

Ksflow must be configured with a cert-manager ClusterIssuer so that it can create Certificates
for KafkaUsers. The ksflow helm chart creates a ClusterRole and ClusterRoleBinding to provide
it with permissions to create/manage those Certificates.

By default, cert-manager approves all Certificates that refer to ClusterIssuers. If your setup
allows more than just ksflow to create cert-manager Certificates (or CertificateRequests), then you may want to put
restrictions in-place for which Certificate/CertificateRequests get created/approved. A couple ways this can be accomplished:
* cert-manager's [approver-policy](https://github.com/cert-manager/approver-policy)
* [kyverno](https://github.com/kyverno/kyverno)

## Kafka Broker Configuration

```properties
ssl.client.auth=required
allow.everyone.if.no.acl.found=false
```
`connections.max.reauth.ms`, can use ACL instead
`ssl.principal.mapping.rules`?
