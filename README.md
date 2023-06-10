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

<!-- <</Stencil::Block>> -->
## Dependencies

### Dependencies

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
## Interacting with Vcluster-Fs-Syncer
