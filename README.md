# OpenTelemetry Collector Kubernetes Example

This repository demonstrates how to configure OpenTelemetry Collector so that logs written to stdout and logs exported via OTLP share nearly identical attributes.

## Install OpenTelemetry Collector

```sh
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm upgrade --install opentelemetry-collector open-telemetry/opentelemetry-collector \
  -f otelcol-values.yaml \
  --namespace opentelemetry-collector \
  --create-namespace
```

This command creates the opentelemetry-collector service with `internalTrafficPolicy: Local`.
The service is reachable at opentelemetry-collector.opentelemetry-collector.svc.cluster.local.

Using this service is essential. Without it, the [Kubernetes attributes processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/k8sattributesprocessor) cannot add pod-level attributes, because telemetry data lacks pod IP address information.


## Run sample application on Minikube

### Build and load image

```sh
docker build . -t example
minikube image load example
```

### Run pod

To ensure that almost the same attributes are assigned to both stdout logs and OTLP-exported logs, configure them using annotations and the `OTEL_RESOURCE_ATTRIBUTES` environment variables:

```sh
kubectl run example-$(date '+%s') --rm --restart=Never --image-pull-policy=Never -i --image=example \
  --annotations="service.name=hello" \
  --annotations="service.version=0.0.1" \
  --env=OTEL_RESOURCE_ATTRIBUTES="service.name=hello,service.version=0.0.1" \
  --env=OTEL_EXPORTER_OTLP_ENDPOINT="http://opentelemetry-collector.opentelemetry-collector.svc.cluster.local:4317"
```

Here are example outputs:

<details>
<summary>Metrics</summary>

```
2025-09-18T16:17:08.889Z        info    Metrics {"resource": {"service.instance.id": "3b8d81d0-711f-45b9-9f18-07a7c6689b66", "service.name": "otelcol-k8s", "service.version": "0.135.0"}, "otelcol.component.id": "debug", "otelcol.component.kind": "exporter", "otelcol.signal": "metrics", "resource metrics": 1, "metrics": 1, "data points": 1}
2025-09-18T16:17:08.889Z        info    ResourceMetrics #0
Resource SchemaURL: https://opentelemetry.io/schemas/1.37.0
Resource attributes:
     -> service.name: Str(hello)
     -> service.version: Str(0.0.1)
     -> telemetry.sdk.language: Str(go)
     -> telemetry.sdk.name: Str(opentelemetry)
     -> telemetry.sdk.version: Str(1.38.0)
     -> k8s.pod.ip: Str(10.244.0.28)
     -> k8s.node.name: Str(minikube)
     -> k8s.pod.name: Str(example-1758212225)
     -> k8s.namespace.name: Str(default)
     -> k8s.pod.start_time: Str(2025-09-18T16:17:07Z)
     -> k8s.pod.uid: Str(37bedbd3-d248-4bf0-9247-ba7f8c40e804)
ScopeMetrics #0
ScopeMetrics SchemaURL:
InstrumentationScope github.com/abicky/opentelemetry-collector-k8s-example
Metric #0
Descriptor:
     -> Name: hello.invocations
     -> Description: The number of invocations
     -> Unit: {invocation}
     -> DataType: Sum
     -> IsMonotonic: true
     -> AggregationTemporality: Cumulative
NumberDataPoints #0
Data point attributes:
     -> key1: Str(value1)
StartTimestamp: 2025-09-18 16:17:08.737929702 +0000 UTC
Timestamp: 2025-09-18 16:17:08.743136869 +0000 UTC
Value: 1
Exemplars:
Exemplar #0
     -> Trace ID: 49d2561d98d3baea5a8ba0686d004462
     -> Span ID: 2379e44179158668
     -> Timestamp: 2025-09-18 16:17:08.738150244 +0000 UTC
     -> Value: 1
        {"resource": {"service.instance.id": "3b8d81d0-711f-45b9-9f18-07a7c6689b66", "service.name": "otelcol-k8s", "service.version": "0.135.0"}, "otelcol.component.id": "debug", "otelcol.component.kind": "exporter", "otelcol.signal": "metrics"}
```

</details>

<details>
<summary>Traces</summary>

```
2025-09-18T16:17:08.889Z        info    Traces  {"resource": {"service.instance.id": "3b8d81d0-711f-45b9-9f18-07a7c6689b66", "service.name": "otelcol-k8s", "service.version": "0.135.0"}, "otelcol.component.id": "debug", "otelcol.component.kind": "exporter", "otelcol.signal": "traces", "resource spans": 1, "spans": 1}
2025-09-18T16:17:08.889Z        info    ResourceSpans #0
Resource SchemaURL: https://opentelemetry.io/schemas/1.37.0
Resource attributes:
     -> service.name: Str(hello)
     -> service.version: Str(0.0.1)
     -> telemetry.sdk.language: Str(go)
     -> telemetry.sdk.name: Str(opentelemetry)
     -> telemetry.sdk.version: Str(1.38.0)
     -> k8s.pod.ip: Str(10.244.0.28)
     -> k8s.node.name: Str(minikube)
     -> k8s.pod.name: Str(example-1758212225)
     -> k8s.namespace.name: Str(default)
     -> k8s.pod.start_time: Str(2025-09-18T16:17:07Z)
     -> k8s.pod.uid: Str(37bedbd3-d248-4bf0-9247-ba7f8c40e804)
ScopeSpans #0
ScopeSpans SchemaURL:
InstrumentationScope github.com/abicky/opentelemetry-collector-k8s-example
Span #0
    Trace ID       : 49d2561d98d3baea5a8ba0686d004462
    Parent ID      :
    ID             : 2379e44179158668
    Name           : run
    Kind           : Internal
    Start time     : 2025-09-18 16:17:08.738126036 +0000 UTC
    End time       : 2025-09-18 16:17:08.738239411 +0000 UTC
    Status code    : Unset
    Status message :
Attributes:
     -> key1: Str(value1)
        {"resource": {"service.instance.id": "3b8d81d0-711f-45b9-9f18-07a7c6689b66", "service.name": "otelcol-k8s", "service.version": "0.135.0"}, "otelcol.component.id": "debug", "otelcol.component.kind": "exporter", "otelcol.signal": "traces"}
```

</details>

<details>
<summary>Logs (OTLP)</summary>

```
2025-09-18T16:17:08.889Z        info    Logs    {"resource": {"service.instance.id": "3b8d81d0-711f-45b9-9f18-07a7c6689b66", "service.name": "otelcol-k8s", "service.version": "0.135.0"}, "otelcol.component.id": "debug", "otelcol.component.kind": "exporter", "otelcol.signal": "logs", "resource logs": 1, "log records": 1}
2025-09-18T16:17:08.890Z        info    ResourceLog #0
Resource SchemaURL: https://opentelemetry.io/schemas/1.37.0
Resource attributes:
     -> service.name: Str(hello)
     -> service.version: Str(0.0.1)
     -> telemetry.sdk.language: Str(go)
     -> telemetry.sdk.name: Str(opentelemetry)
     -> telemetry.sdk.version: Str(1.38.0)
     -> k8s.pod.ip: Str(10.244.0.28)
     -> k8s.node.name: Str(minikube)
     -> k8s.pod.name: Str(example-1758212225)
     -> k8s.namespace.name: Str(default)
     -> k8s.pod.start_time: Str(2025-09-18T16:17:07Z)
     -> k8s.pod.uid: Str(37bedbd3-d248-4bf0-9247-ba7f8c40e804)
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope github.com/abicky/opentelemetry-collector-k8s-example
LogRecord #0
ObservedTimestamp: 2025-09-18 16:17:08.738208661 +0000 UTC
Timestamp: 2025-09-18 16:17:08.738192036 +0000 UTC
SeverityText: INFO
SeverityNumber: Info(9)
Body: Str(Hello World!)
Attributes:
     -> key1: Str(value1)
Trace ID: 49d2561d98d3baea5a8ba0686d004462
Span ID: 2379e44179158668
Flags: 1
        {"resource": {"service.instance.id": "3b8d81d0-711f-45b9-9f18-07a7c6689b66", "service.name": "otelcol-k8s", "service.version": "0.135.0"}, "otelcol.component.id": "debug", "otelcol.component.kind": "exporter", "otelcol.signal": "logs"}
```

</details>

<details>
<summary>Logs (stdout)</summary>

```
2025-09-18T16:17:09.091Z        info    Logs    {"resource": {"service.instance.id": "3b8d81d0-711f-45b9-9f18-07a7c6689b66", "service.name": "otelcol-k8s", "service.version": "0.135.0"}, "otelcol.component.id": "debug", "otelcol.component.kind": "exporter", "otelcol.signal": "logs", "resource logs": 3, "log records": 38}
2025-09-18T16:17:09.091Z        info    ResourceLog #0
Resource SchemaURL:
Resource attributes:
     -> k8s.container.restart_count: Str(0)
     -> k8s.pod.uid: Str(37bedbd3-d248-4bf0-9247-ba7f8c40e804)
     -> k8s.container.name: Str(example-1758212225)
     -> k8s.namespace.name: Str(default)
     -> k8s.pod.name: Str(example-1758212225)
     -> k8s.pod.start_time: Str(2025-09-18T16:17:07Z)
     -> k8s.node.name: Str(minikube)
     -> service.name: Str(hello)
     -> service.version: Str(0.0.1)
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope
LogRecord #0
ObservedTimestamp: 2025-09-18 16:17:08.865213244 +0000 UTC
Timestamp: 2025-09-18 16:17:08.738902036 +0000 UTC
SeverityText:
SeverityNumber: Unspecified(0)
Body: Str([INFO] Hello World!]
)
Attributes:
     -> log.iostream: Str(stdout)
     -> log.file.path: Str(/var/log/pods/default_example-1758212225_37bedbd3-d248-4bf0-9247-ba7f8c40e804/example-1758212225/0.log)
Trace ID:
Span ID:
Flags: 0
```

</details>

As you can see, all the telemetry data has the `service.name`, `service.version`, `k8s.node.name`, `k8s.pod.name`, `k8s.namespace.name`, `k8s.pod.start_time`, and `k8s.pod.uid` attributes.

## Why not use Node IP?

You might consider exporting directly to the node IP as follows:


```sh
kubectl run example-$(date '+%s') --rm --restart=Never -i --image=example \
  --annotations="service.name=hello" \
  --annotations="service.version=0.0.1" \
  --overrides='{
  "spec": {
    "containers": [
      {
        "name": "example",
        "image": "example",
        "imagePullPolicy": "Never",
        "env": [
          {
            "name": "K8S_NODE_IP",
            "valueFrom": {
              "fieldRef": {
                "fieldPath": "status.hostIP"
              }
            }
          },
          {
            "name": "OTEL_RESOURCE_ATTRIBUTES",
            "value": "service.name=hello,service.version=0.0.1"
          },
          {
            "name": "OTEL_EXPORTER_OTLP_ENDPOINT",
            "value": "http://$(K8S_NODE_IP):4317"
          }
        ]
      }
    ]
  }
}'
```

However, this does not work as expected. Since the source IP (pod IP) is masqueraded, the Kubernetes attributes processor cannot [map the IP back to a pod](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.135.0/processor/k8sattributesprocessor/processor.go#L158). As a result, pod-level attributes are missing.


### Debug OpenTelemetry Collector

First, build the otelcol Docker image:

```sh
cd otelcol
make docker-image
minikube image load otelcol:0.0.1
```

Then, patch the opentelemetry-collector-agent daemonset to use the new image:

```sh
kubectl patch ds -n opentelemetry-collector opentelemetry-collector-agent -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "opentelemetry-collector",
          "image": "otelcol:0.0.1",
          "livenessProbe": null,
          "readinessProbe": null
        }]
      }
    }
  }
}'
```

Once port-forwarding 2345 port, you can connect the OpenTelemetry Collector process usign your debugger:

```sh
kubectl port-forward -n opentelemetry-collector ds/opentelemetry-collector-agent 2345:2345
```
