# Nats Source
Nats Source is a user-defined source for [Numaflow](https://numaflow.numaproj.io/)
that facilitates reading messages from a Nats server.

- [Quick Start](#Quick-Start)
- [Using Nats Source in Your Numaflow Pipeline](#how-to-use-the-nats-source-in-our-own-numaflow-pipeline)
- [JSON Configuration](#using-json-format-to-specify-the-nats-source-configuration)

## Quick Start
This quick start guide will help you to set up and run a Nats source in a Numaflow pipeline on your local kube cluster. Follow the steps below to get started:

### Prerequisites
* [Install Numaflow on your local kube cluster](https://numaflow.numaproj.io/quick-start/)
* [Install The Nats CLI tool](https://github.com/nats-io/natscli)

### Step-by-step Guide

#### 1. Deploy a Nats Server and a Numaflow Pipeline

In the current folder, run:
```bash
kubectl apply -k ./example
```

#### 2. Verify the Pipeline

Execute the following command to verify the pipeline is up and running:
```bash
kubectl get pipeline nats-source-e2e
```
You should see:
```
NAME              PHASE     MESSAGE   VERTICES   AGE
nats-source-e2e   Running             3          1m
```
#### 3. Send Messages to the Nats server

Port-forward the Nats server to your local machine:
```bash
kubectl port-forward svc/nats 4222:4222
```

Next, send messages:
```bash
nats pub test-subject "Hello World" --user=testingtoken
```

#### 4. Verify the Log Sink

Replace the "xxxxx" with the appropriate out vertex pod name:
```bash
kubectl logs nats-source-e2e-out-0-xxxxx
```

You should see the output similar to:
```
2023/09/05 19:18:44 (out)  Payload -  Hello World  Keys -  []  EventTime -  1693941455870
```

#### 5. Clean up

To delete the Numaflow pipeline and the Nats server, run:
```bash
kubectl delete -k ./example
```

Congratulations! You have successfully run a Nats source in a Numaflow pipeline on your local kube cluster.

## How to use the Nats source in our own Numaflow pipeline

To integrate the Nats source in your own Numaflow pipeline, follow these detailed steps:

### 1. Deploy your Nats server
Deploy our own Nats server to our cluster. Refer to the [NATS Docs](https://docs.nats.io/running-a-nats-service/introduction) for guidance.

### 2: Create a ConfigMap
Define the Nats source configuration in a ConfigMap and mount it to the Nats source pod as a volume. Create a ConfigMap using the example below:

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

### 3. Specify the Nats Source in the Pipeline
Name your Nats Configuration ConfigMap as `nats-config.yaml` and mount it to the Nats source pod as a volume under path `/etc/config`.
Create all the secrets that are referenced in the Nats source configuration and mount them to the Nats source pod as volumes under path `/etc/secrets/{secret-name}`.

Include the Nats Source in your pipeline using the template below:

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

Here is a template for creating the secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: nats-auth-fake-token
stringData:
  fake-token: "testingtoken"
```

### 4: Run the Pipeline
Now, execute the pipeline to start reading messages from the Nats server.

## Using JSON format to specify the Nats source configuration
By default, Numaflow Nats Source uses YAML as configuration format.

You can also specify the Nats source configuration in JSON format. Find below a guide on how to set the configuration using JSON:

### ConfigMap in JSON
Here is how you can craft a ConfigMap in JSON format:
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

### Pipeline Template Adjustment
Adjust your pipeline template to facilitate JSON configuration as shown below:
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

Remember to set the `CONFIG_FORMAT` environment variable to `json`.