## Quick Start

#### Prerequisites
* Kubernetes (i.e. [k3d](https://k3d.io/v5.4.6/#installation), [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation), etc.)
* [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
* [helm](https://helm.sh/docs/intro/install/)
* [openssl](https://github.com/openssl/openssl) (there is an [open issue for name-constraints support](https://github.com/cert-manager/cert-manager/issues/3655) in cert-manager)

#### Generate Certificates
```shell
mkdir /tmp/ksflow-quickstart-certs
cd /tmp/ksflow-quickstart-certs

####### CA
# generate private key for CA (ksflow-quickstart-root-ca.key)
openssl genrsa -out ksflow-quickstart-root-ca.key 4096
# use private key to generate self-signed CA (ksflow-quickstart-root-ca.crt)
openssl req -new -x509 -nodes \
  -key ksflow-quickstart-root-ca.key \
  -out ksflow-quickstart-root-ca.crt \
  -subj "/CN=KsflowQuickstartCA-Root"

####### Kafka
# generate a private key for the kafka server certificate
openssl genrsa -out ksflow-quickstart-kafka.key 4096
# generate a certificate signing request (CSR)
openssl req -new \
  -key ksflow-quickstart-kafka.key \
  -out ksflow-quickstart-kafka.csr \
  -subj "/CN=broker1.ksflow-quickstart-kafka.svc.cluster.local"
# generate the server certificate by signing the CSR with the CA certificate
openssl x509 -req \
  -in ksflow-quickstart-kafka.csr \
  -CA ksflow-quickstart-root-ca.crt \
  -CAkey ksflow-quickstart-root-ca.key \
  -CAcreateserial \
  -out ksflow-quickstart-kafka.crt \
  -days 1

####### Ksflow
# create config file with name-constraints
cat <<EOF > ksflow-quickstart-controller-ca.conf
[ req ]
distinguished_name = req_distinguished_name
x509_extensions = v3_ca

[ req_distinguished_name ]
commonName             = KsflowQuickstartControllerCA-Intermediate
countryName            = AB
stateOrProvinceName    = CD
localityName           = EFG_HIJ
0.organizationName     = MyOrg
organizationalUnitName = MyOrgUnit
emailAddress           = myemail@example.com

[ v3_ca ]
basicConstraints = critical,CA:true,pathlen:0
nameConstraints = critical,permitted;DNS:.ksflow.ksflow-quickstart.example.com
EOF
# generate a private key and certificate for CA
openssl genrsa -out ksflow-quickstart-controller-ca.key 4096
openssl req -new \
  -key ksflow-quickstart-controller-ca.key \
  -out ksflow-quickstart-controller-ca.csr \
  -subj "/CN=KsflowQuickstartControllerCA-Intermediate" \
  -config ksflow-quickstart-controller-ca.conf
openssl x509 -req \
  -in ksflow-quickstart-controller-ca.csr \
  -CA ksflow-quickstart-root-ca.crt \
  -CAkey ksflow-quickstart-root-ca.key \
  -CAcreateserial \
  -out ksflow-quickstart-controller-ca.crt \
  -days 1 \
  -extfile ksflow-quickstart-controller-ca.conf \
  -extensions v3_ca
```

#### Install
```shell
# install kafka
kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-kafka.yaml

# install ksflow
helm repo add https://ksflow.github.io/ksflow-helm
helm upgrade ksflow ksflow/ksflow --install \
  --values https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-values.yaml
```

#### Create a Topic
```shell
# create a KafkaTopic
kubectl apply -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-kt.yaml

# watch the KafkaTopic until it's STATUS is "Available"
kubectl get kt --watch

# verify the topic was created in kafka
kubectl exec deploy/ksflow-quickstart-kafka -- /bin/sh -c "/opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list"
```

#### Delete the Topic
```shell
# delete the KafkaTopic
kubectl delete kt quickstart

# verify the topic was deleted from kafka
kubectl exec deploy/ksflow-quickstart-kafka -- /bin/sh -c "/opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list"
```

#### Uninstall
```shell
kubectl delete -f https://raw.githubusercontent.com/ksflow/ksflow/main/config/samples/quickstart-kafka.yaml
helm uninstall ksflow
```
