---
title: "TLS/SSL connections"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 1600
---
Each cluster created is automatically equipped with a self-generated TLS certificate and is preconfigured for the use of TLS/SSL. However, this certificate is not based on a Certificate Authority (CA) that is known to the clients. This means that although communication between the client and server is encrypted, the certificate cannot be verified by the client.

The following chapter deals with the creation of custom certificates and the steps required to integrate these certificates into the PostgreSQL cluster. In the example, a custom CA is created, on the basis of which the certificates are then generated and signed by this CA. This step can be skipped if certificates have already been obtained from another trusted organisation.

### Create a custom CA and Certificates 
{{< hint type=important >}} Precondition: This chapter requires openssl {{< /hint >}}
#### Create the CA
The first step is to create a custom CA. An organisation name is required for this. You can also add further details about the country, district and location.
The CA serves as the central authority that signs the certificates and thus guarantees the correctness of the certificate. In order to successfully complete the verification of a certificate, the CA's certificate must be stored on the client system. 
```
ORGANIZATION=MyCustomOrganization
CA=$ORGANIZATION-RootCA

mkdir $CA
cd $CA

# Creating the CA-Key
openssl genpkey -algorithm EC -out $CA.key -pkeyopt ec_paramgen_curve:secp384r1 -pkeyopt ec_param_enc:named_curve -aes256

# Creating the CA-Certificate
openssl req -x509 -new -nodes -key $CA.key -sha512 -days 1826 -out $CA.crt -subj "/CN=${ORGANIZATION} Root-CA/C=AT/ST=Lower Austria/L=Woellersdorf/O=${ORGANIZATION}"

```

#### Create a custom Certificate
The server needs a certificate signed by a CA and a private key so that it can claim to be trustworthy. 

{{< hint type=important >}} It is important that the CA certificate is stored as trustworthy with the client. Otherwise, no certificate check is possible. {{< /hint >}}


```
CN=cluster-1
DNS2="${CN}-repl"
DNS3="${CN}-pooler"
DNS4="${CN}-pooler-repl"

# Creating the private Key
openssl genpkey -algorithm EC -out $CN.key -pkeyopt ec_paramgen_curve:secp384r1 -pkeyopt ec_param_enc:named_curve

# Creating Certificate Signing Request (CSR))
openssl req -new -key $CN.key -out $CN.csr \
  -subj "/C=AT/ST=Lower Austria/L=Woellersdorf/O=${ORGANIZATION}/OU=OrgUnit/CN=${CN}" \
  -addext "subjectAltName=DNS:${CN},DNS:${DNS2},DNS:${DNS3},DNS:${DNS4}"


# Sign CSR with the CA
openssl x509 -req -in $CN.csr -CA $CA.crt -CAkey $CA.key -CAcreateserial -out $CN.crt -days 365 \
  -extfile <(echo -e "[ v3_req ]\nsubjectAltName=DNS:${CN},DNS:${DNS2},DNS:${DNS3},DNS:${DNS4}") -extensions v3_req

```

#### Add Certicate to the Cluster

For adding the Certificate to your cluster a secret on kubernetes is needed.
There are two different options here. 
For the first option, a secret is created that contains all the necessary information. I.e. 
- Server certificate
- Private server key
- CA certificate
In the second variant, the CA certificate is separated and written in a separate secret. The advantage of this is that the CA only needs to be saved once and changed in the event of an update. 

##### First Option: Using one secret for all three objects

```
kubectl create secret generic cluster-1-tls \
  --from-file=tls.crt=$CN.crt \
  --from-file=tls.key=$CN.key \
  --from-file=ca.crt=$CA.crt
```

Finally, the definition is made in the cluster manifest so that the operator adapts the cluster. 

```yaml
apiVersion: "cpo.opensource.cybertec.at/v1"
kind: postgresql
...
metadata:
  name: cluster-1
spec:
  tls:
    secretName: "cluster-1-tls"
    caFile: "ca.crt"
```

##### Second Option: Using a separat Secret for the CA

```
kubectl create secret generic cpo-root-ca --from-file=ca.crt=ca.crt
```

```
kubectl create secret generic cluster-1-tls \
  --from-file=tls.crt=$CN.crt \
  --from-file=tls.key=$CN.key \
```

Finally, the definition is made in the cluster manifest so that the operator adapts the cluster. 

```yaml
apiVersion: "cpo.opensource.cybertec.at/v1"
kind: postgresql

metadata:
  name: cluster-1
spec:
  tls:
    secretName: "cluster-1-tls"
    caSecretName: "cpo-root-ca"
    caFile: "ca.crt"
```

A regular check of the mounted certificates takes place automatically within the container. This check takes place every 5 minutes. If the certificates have been updated, the certificates are loaded automatically.

{{< hint type=important >}} In addition to generating the certificates independently, [cert-manager](https://cert-manager.io/docs/) can also be used for this purpose.  {{< /hint >}}
