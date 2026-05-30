# Label Sync Operator

A Kubernetes Operator built with Kubebuilder and Go that automatically synchronizes labels across Kubernetes resources to ensure consistency and simplify resource management.

## Overview

Managing labels across multiple Kubernetes resources can become difficult as applications grow. Inconsistent labels can impact resource selection, monitoring, network policies, RBAC configurations, and operational workflows.

The Label Sync Operator automates label synchronization by ensuring that target resources always contain the desired labels defined in a custom resource.

## Features

* Automatic label synchronization
* Continuous reconciliation of label drift
* Supports Deployments, Services, ConfigMaps, and Secrets
* Ensures consistent labeling across resources
* Kubernetes-native implementation using controller-runtime
* Declarative management through Custom Resources

## Architecture

```text
LabelSync CR
      |
      v
Label Sync Operator
      |
      v
Target Resources
(Deployments, Services,
 ConfigMaps, Secrets)
      |
      v
Labels Applied / Updated
```

## Custom Resource Example

```yaml
apiVersion: apps.dev.com/v1alpha1
kind: LabelSync
metadata:
  name: sample-labelsync
spec:
  targetNamespace: default
  labels:
    environment: production
    team: platform
    managed-by: label-sync-operator
```

## How It Works

1. User creates a LabelSync custom resource.
2. Operator watches the custom resource.
3. Operator discovers matching resources.
4. Desired labels are compared with existing labels.
5. Missing or outdated labels are synchronized.
6. Operator continuously reconciles drift.

## Use Cases

* Standardizing application labels
* Multi-team Kubernetes environments
* Cost allocation and chargeback
* Monitoring and observability tagging
* Policy enforcement
* Resource governance

## Technologies Used

* Go
* Kubernetes
* Kubebuilder
* controller-runtime
* Custom Resource Definitions (CRDs)

## Future Enhancements

* Label inheritance
* Namespace-wide synchronization
* Resource selector support
* Annotation synchronization
* Webhook validation
* Metrics and Prometheus integration

## Learning Outcomes

This project demonstrates:

* Kubernetes Operator Development
* Reconciliation Patterns
* Resource Discovery
* Label Management
* Kubernetes API Interactions
* Owner References
* Watches and Predicates
* Controller Runtime

```
```
