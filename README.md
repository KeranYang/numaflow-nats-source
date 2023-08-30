# Nats Source
A simple example of a Nats source.

## Use it in numaflow e2e tests
This example demonstrates how to use a Nats source in numaflow e2e tests.

### Step 1: Deploy Nats app to your cluster under numaflow-system namespace
Go to Numaflow workspace, under `numaflow/test/nats-e2e` folder, run:
```bash
kubectl -n numaflow-system delete statefulset nats --ignore-not-found=true
kubectl apply -k ../../config/apps/nats -n numaflow-system
```
It will start a Nats server.

### Step 2: Prepare the Nats source configuration
A Nats configuration specifies information required to connect to a Nats server.
The configuration is stored in a ConfigMap, which is mounted to the Nats source pod as a volume.
The following example demonstrates how to configure a Nats source to connect to a Nats server with authentication token.

Go to `numaflow/test/nats-e2e/testdata` folder

1. Create the authentication token secret, create a file named `nats-auth-fake-token` with the following content:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: nats-source-fake-token
stringData:
  fake-token: "testingtoken"
```
Note:
The secret token value is the same as the one
we declared in our Nats app deployment file `numaflow/config/apps/nats/nats.yaml`,
such that the e2e test can connect to the Nats server.

1. Create the ConfigMap, create a file named `nats-source-config.yaml` with the following content:
```yaml
apiVersion: v1
data:
  nats-config.json: |
    {
         "url":"nats",
         "subject":"test-subject",
         "queue":"my-queue",
         "auth":{
            "token":{
               "name":"nats-auth-fake-token",
               "key":"fake-token"
            }
         }
      }
kind: ConfigMap
metadata:
  name: nats-config-map
```

1. Deploy the secret and ConfigMap to your cluster under numaflow-system namespace:
```bash
kubectl apply -f testdata/nats-auth-fake-token.yaml -n numaflow-system
kubectl apply -f testdata/nats-source-config.yaml -n numaflow-system
```

Up until now, we have a running Nats server, and a ConfigMap that contains the Nats source configuration.
We also have a secret that contains the authentication token for the Nats server.
Time to specify the pipeline.

### Step 3: Specify the pipeline
* Build the Nats source image, push it to a public registry, and use it as the user-defined source image.

With in this repo, run the following commands:
```bash
make image
docker push quay.io/numaio/numaflow-go/keran-test-nats-source:secret0.5.0 
```
* Mount the ConfigMap to the Nats source pod as a volume.

```yaml
apiVersion: numaflow.numaproj.io/v1alpha1
kind: Pipeline
metadata:
  name: nats-source-e2e
spec:
  vertices:
    - name: in
      scale:
        min: 2
      volumes:
        - name: my-config-mount
          configMap:
            name: nats-config-map
        - name: my-secret-mount
          secret:
            secretName: nats-auth-fake-token
      source:
        udsource:
          container:
            image: quay.io/numaio/numaflow-go/keran-test-nats-source:secret0.5.1
            volumeMounts:
              - name: my-config-mount
                mountPath: /etc/config
              - name: my-secret-mount
                mountPath: /etc/secrets/nats-auth-fake-token
    - name: p1
      udf:
        builtin:
          name: cat
    - name: out
      sink:
        log: {}
  edges:
    - from: in
      to: p1
    - from: p1
      to: out
```

### Step 4: Run the e2e test
* Go to numaflow root folder, run:
```bash
make TestNatsSource
```