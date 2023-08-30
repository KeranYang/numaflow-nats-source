# Nats Source
A simple example of a Nats source.

## Use it in numaflow e2e tests

### Step 1: Deploy Nats app to your cluster
Go to Numaflow workspace, under `numaflow/test/nats-e2e` folder, run:
```bash
kubectl -n numaflow-system delete statefulset nats --ignore-not-found=true
kubectl apply -k ../../config/apps/nats -n numaflow-system
```
This will start a Nats server in your cluster under numaflow-system namespace.

### Step 2: Prepare the Nats source config

#### Option One: hard code authentication tokens
* Use the `main` branch of this repo to build the Nats source image.
* Go to `numaflow/test/nats-e2e/testdata` folder
** Create a file named `nats-source-config.yaml` with the following content:
```yaml
apiVersion: v1
data:
  nats-config.json: "{\n\t\"url\": \"nats\",\n\t\"subject\": \"test-subject\",\n\t\"queue\":
    \"my-queue\",\n\t\"auth\": {\n\t\t\"token\": \"testingtoken\"\n\t}\n}\n"
kind: ConfigMap
metadata:
  name: nats-config-map
```
* This will create a ConfigMap named `nats-config-map` with a file named `nats-config.json` in it.
* The `nats-config.json` file contains the Nats source configuration.
* The `token` field is the authentication token for Nats server.
* Deploy the ConfigMap to your cluster:
```bash
kubectl apply -f testdata/nats-source-config.yaml -n numaflow-system
```

#### Option Two: use secret to store authentication tokens
* Use the `use-secret` branch of this repo to build the Nats source image.

### Step 3: Specify the pipeline
* Mount the ConfigMap to the Nats source pod as a volume.
* Use the image built in Step 1 as the user-defined source image.
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
        - name: my-file-config
          configMap:
            name: nats-config-map
      source:
        udsource:
          container:
            env:
              - name: abc
                value: |
                  {"a": "b", "c": "d"}
            # A simple user-defined source for e2e testing
            # See https://github.com/numaproj/numaflow-go/tree/main/pkg/sourcer/examples/simple_source
            image: quay.io/numaio/numaflow-go/source-simple-source:v0.5.9
            volumeMounts:
              - name: my-file-config
                mountPath: /etc/config
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