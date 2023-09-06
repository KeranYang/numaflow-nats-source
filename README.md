# Nats Source
Nats Source is a [Numaflow](https://numaflow.numaproj.io/) user-defined source that reads messages from a Nats server.

## Quick Start
Following is a quick start guide to run a Nats source in a Numaflow pipeline hosted on your local kube cluster.
The pipeline reads messages from a Nats server, and writes them to a log sink.

### Pre-requisites
* Follow the [Numaflow quick start guide](https://numaflow.numaproj.io/docs/quickstart) to install Numaflow on your local kube cluster.
* Follow [natscli](https://github.com/nats-io/natscli) to install The Nats Command Line Interface (CLI) tool.

### Step 1: Deploy a Nats server, and a Numaflow pipeline
Under the current folder, run the following command
```bash
kubectl apply -k ./example
```

### Step 2: Verify the pipeline
Run the following command to verify the pipeline is up and running:
```bash
kubectl get pipeline nats-source-e2e
```
The output should be similar to:
```
NAME              PHASE     MESSAGE   VERTICES   AGE
nats-source-e2e   Running             3          1m
```

### Step 3: Send messages to the Nats server
Run the following command to port-forward the Nats server to your local machine:
```bash
kubectl port-forward svc/nats 4222:4222
```
Run the following command to send messages to the Nats server:
```bash
nats pub test-subject "Hello World" --user=testingtoken
```

### Step 4: Verify the log printed by the log sink
Run the command below and remember to replace "xxxxx" with the appropriate out vertex pod name.
```bash
kubectl logs nats-source-e2e-out-0-xxxxx
```
The output should be similar to:
```
2023/09/05 19:18:44 (out)  Payload -  Hello World  Keys -  []  EventTime -  1693941455870
```

### Step 5: Clean up
Run the following command to delete the Numaflow pipeline and the Nats server:
```bash
kubectl delete -k ./example
```

Hurray!
We have successfully run a Nats source in a Numaflow pipeline.
Now let's dive into the details of how to use the Nats source in our own Numaflow pipeline.

## How to use the Nats source in our own Numaflow pipeline

### Step 1: Deploy our own Nats server
Deploy our own Nats server to our cluster. There are multiple ways to do this.
E.g., follow the [NATS Docs](https://docs.nats.io/running-a-nats-service/introduction)

### Step 2: Create a ConfigMap that contains the Nats source configuration
With a running Nats server, we need to specify how to connect to the Nats server.
We need a **Nats source configuration**.
Numaflow Nats Source requires us to specify the Nats source configuration in a ConfigMap,
and mount it to the Nats source pod as a volume.
The following example demonstrates how to create a ConfigMap that contains the Nats source configuration in YAML format.

```yaml
apiVersion: v1
data:
  nats-config.yaml: |
    url: nats
    subject: test-subject
    queue: my-queue
    auth:
      token:
        localobjectreference:
          name: nats-auth-fake-token
        key: fake-token
kind: ConfigMap
metadata:
  name: nats-config-map
```

The configuration contains the following fields:
* `url`: The Nats server URL.
* `subject`: The Nats subject to subscribe to.
* `queue`: The Nats queue group name.
* `auth`: The Nats authentication information.
  * `token`: The Nats authentication token information.
    * `name`: The name of the secret that contains the authentication token.
    * `key`: The key of the authentication token in the secret.

### Step 3: Specify the Nats source in the pipeline
With the Nats source configuration, we can specify the Nats source in the pipeline.
The following example demonstrates how to specify the Nats source in the pipeline.

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
            image: quay.io/numaio/numaflow-source/nats-source:v0.5.2
            volumeMounts:
              - name: my-config-mount
                mountPath: /etc/config
              - name: my-secret-mount
                mountPath: /etc/secrets/nats-auth-fake-token
    - name: out
      sink:
        log: {}
  edges:
    - from: in
      to: out
```
The Nats source is specified in the `in` vertex.
The Nats source is a user-defined source, so we need to specify the user-defined source image.
In this example, we use the Nats source image `quay.io/numaio/numaflow-source/nats-source:v0.5.2`.
We also need to mount the ConfigMap that contains the Nats source configuration to the Nats source pod as a volume.
In this example, we mount the ConfigMap to the Nats source pod as a volume named `my-config-mount`.

Note: The Nats source requires the Nats authentication token to connect to the Nats server.
Hence, we also need to mount the secret that contains the authentication token to the Nats source pod as a volume.
In this example, we mount the secret to the Nats source pod as a volume named `my-secret-mount`.
The following template was used to create the secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: nats-auth-fake-token
stringData:
  fake-token: "testingtoken"
```

### Step 4: Run the pipeline
With the steps above, we have created a running Nats server, specified the Nats source configuration in a ConfigMap,
and specified the Nats source in the pipeline template.

Now we can run the pipeline and start reading messages from the Nats server.

## Using JSON format to specify the Nats source configuration
By default, Numaflow Nats Source uses YAML as configuration format.
You can also specify the Nats source configuration in a ConfigMap in JSON format.
You can tell the Nats source to read the Nats source configuration in YAML format by setting the environment variable `NATS_CONFIG_FORMAT` to `json`.
The following example demonstrates how to specify the Nats source configuration in JSON format.

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
The pipeline template remains the same except that we need to set the environment variable `CONFIG_FORMAT` to `json`.
```yaml
source:
  udsource:
    container:
      image: quay.io/numaio/numaflow-source/nats-source:v0.5.2
      env:
        - name: CONFIG_FORMAT
          value: json
      volumeMounts:
        ...
```