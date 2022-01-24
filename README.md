# Step Pod Autoscaler

Scales deployments in steps using HPA.

![image](https://user-images.githubusercontent.com/19834040/150770081-206f2fa0-e1b0-4ba8-acab-5295a85bb983.png)

Refer [controller/*controller.go](https://github.com/danishprakash/kube-step-podautoscaler/blob/main/controllers/deployment_controller.go) for implementation details and explanation for a better understanding.

### Example Run

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
