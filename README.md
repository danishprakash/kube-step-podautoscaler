# Step Pod Autoscaler

Scales deployments in steps using HPA.

For a config of kind:
```
deployment: nodeservice
goalReplicas: 10
stepSize: 2
```

The controller would operate like:
```text
1.643009   INFO    controller.deployment   Fetched HPA, calculating desiredReplicas                        {"reconciler kind": "Deployment", "name": "nodeservice", "HPA": "nodeservice"}
1.6430+09  INFO    controller.deployment   Updating HPA                                                    {"reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "HPA": "nodeservice", "min/max": 10}
1.6430+09  INFO    controller.deployment   HPA resourceVersion didn't match, requeuing request             { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default"}
1.6430+09  INFO    controller.deployment   Deployment is waiting for pods to come up, skipping updates.    { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "deployment": "nodeservice"}
1.6430+09  INFO    controller.deployment   Deployment is waiting for pods to come up, skipping updates.    { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "deployment": "nodeservice"}
1.6430+09  INFO    controller.deployment   Deployment is waiting for pods to come up, skipping updates.    { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "deployment": "nodeservice"}
1.6430+09  INFO    controller.deployment   Deployment is waiting for pods to come up, skipping updates.    { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "deployment": "nodeservice"}
1.6430+09  INFO    controller.deployment   Deployment is waiting for pods to come up, skipping updates.    { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "deployment": "nodeservice"}
1.643009   INFO    controller.deployment   Fetched HPA, calculating desiredReplicas                        { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "HPA": "nodeservice"}
1.6430+09  INFO    controller.deployment   Updating HPA                                                    { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "HPA": "nodeservice", "min/max": 10}
1.6430+09  INFO    controller.deployment   Fetched HPA, calculating desiredReplicas                        { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "HPA": "nodeservice"}
1.6430+09  INFO    controller.deployment   Updating HPA                                                    { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "HPA": "nodeservice", "min/max": 10
1.6430+09  INFO    controller.deployment   Deployment is waiting for pods to come up, skipping updates.    { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "deployment": "nodeservice"}
1.6430+09  INFO    controller.deployment   Deployment is waiting for pods to come up, skipping updates.    { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "deployment": "nodeservice"}
1.6430+09  INFO    controller.deployment   Deployment up to date, skipping further updates.                { "reconciler kind": "Deployment", "name": "nodeservice", "namespace": "default", "deployment": "nodeservice"}
```
