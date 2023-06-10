# vcluster-fs-syncer

<!-- <<Stencil::Block(customGeneralInformation)>> -->

<!-- <</Stencil::Block>> -->

## Prerequisites

<!-- <<Stencil::Block(customPrerequisites)>> -->

<!-- <</Stencil::Block>> -->

## Building and Testing

This project uses devbase, which exposes the following build tooling: [devbase/docs/makefile.md](https://github.com/getoutreach/devbase/blob/main/docs/makefile.md)

<!-- <<Stencil::Block(customBuildingAndTesting)>> -->

<!-- <</Stencil::Block>> -->
### Building and Running

If you want to add this to your developer environment, please check out the section in the
README.md about [adding to this developer environment](https://github.com/getoutreach/vcluster-fs-syncer#add-to-your-development-environment).

If you want to run this locally, you can do the following:

```bash
devenv provision
devenv tunnel
```

and in a separate terminal, since `devenv tunnel` is a blocking operation, run the following
in the root of this repository:

```bash
make devserver
```

This will build and run your project locally, using the developer environment to provide any
integrations and dependent services that are tunneled to your local network.

### Generating Deployment Manifests Locally

If you want to observe the deployment manifests generated when running the service in the developer
environment, you can leverage the following script:

```bash
./scripts/shell-wrapper.sh deploy-to-dev.sh show
```

### Replacing a Remote Version of the a Package with Local Version

_This is only applicable if this repository exposes a public package_.

If you want to test a package exposed in this repository in a project that uses it, you can
add the following `replace` directive to that project's `go.mod` file:

```
replace github.com/getoutreach/vcluster-fs-syncer => /path/to/local/version/vcluster-fs-syncer
```

**_Note_**: This repository may have postfixed it's module path with a version, go check the first
line of the `go.mod` file in this repository to see if that is the case. If that is the case,
you will need to modify the first part of the replace directive (the part before the `=>`) with
that postfixed path.

### Linting and Unit Testing

You can run the linters and unit tests with:

```bash
make test
```
### End-to-end Testing

You can run end-to-end tests with:

```bash
make e2e
```

This leverages the developer environment to interact with dependent integrations and services. If
an already provisioned environment exists it will use that, else it will create one.
