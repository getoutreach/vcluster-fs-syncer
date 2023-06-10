# vcluster-fs-syncer
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/getoutreach/vcluster-fs-syncer)
[![Generated via Bootstrap](https://img.shields.io/badge/Outreach-Bootstrap-%235951ff)](https://github.com/getoutreach/bootstrap)
[![Coverage Status](https://coveralls.io/repos/github/getoutreach/vcluster-fs-syncer/badge.svg?branch=main)](https://coveralls.io/github//getoutreach/vcluster-fs-syncer?branch=main)
<!-- <<Stencil::Block(extraBadges)>> -->

<!-- <</Stencil::Block>> -->

Synchronizes `/var/lib/kubelet/pods` into predictable vcluster scoped paths  at `/var/lib/loft/<vclusterName>/pods`

## Contributing

Please read the [CONTRIBUTING.md](CONTRIBUTING.md) document for guidelines on developing and contributing changes.

## High-level Overview

<!-- <<Stencil::Block(overview)>> -->
This service is a light-weight service that reacts to pod creation, update, and delete events.

When a pod event is received, it looks for the following annotations to determine if a pod is deployed in a vcluster
and to grab information on the virtual pod (within the vcluster):

* `vcluster.loft.sh/uid` - UUID of the pod inside the vcluster
* `vcluster.loft.sh/name` - Name of the pod inside the vcluster
* `vcluster.loft.sh/namespace` - Namespace of the pod inside the vcluster
* `vcluster.loft.sh/managed-by` - Name of the vcluster that the pod is deployed in

If the pod contains all of these annotations, the pod is considered to be deployed in a vcluster and the [pod directory](https://yuminlee2.medium.com/kubernetes-folder-structure-and-functionality-overview-5b4ec10c32bf)
is bind mounted into the vcluster scoped directory at `/var/lib/loft/<vclusterName>/pods/<podUID>`.

If the pod does not contain all of these annotations, the pod is considered to be deployed in the host cluster and no
action is taken.

The ideal use case for this service is to run daemonsets inside of a virtual cluster that need access to the pod
directory. For example, Velero which needs it for the `restic` (now called `node-agent`) daemonset. In order to use
the per-vcluster pod directory, the `/var/lib/loft/<vclusterName>/pods` directory must be mounted in your Kubernetes
resource at whatever path you need it to be.

### Deploying

The vcluster-fs-syncer is deployed as a Kubernetes Deployment. It is deployed in the `vcluster-fs-syncer--bento1a` namespace.

To deploy, run the following:

```bash
./scripts/shell-wrapper.sh deploy-to-dev.sh show | kubectl apply -f -
```

<!-- <</Stencil::Block>> -->

### Adding and Deleting Service in Development Environment

First, make sure you [set up your development environment](https://github.com/getoutreach/devenv#getting-started).

To add this service to your developer environment:

```bash
devenv apps deploy vcluster-fs-syncer
```

To delete this service from your developer environment:

```bash
devenv apps delete vcluster-fs-syncer
```
