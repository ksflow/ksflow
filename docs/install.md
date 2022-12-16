## Install

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