# vcluster-fs-syncer

<!--- Block(custom) -->
<!--
We expect CONTRIBUTING.md to look mostly identical for all bootstrap services.

If your service requires special instructions for developers, you can place
those instructions in this block. If your service isn't special, it's safe to
leave this comment here as-is.

If the text you are about to add here applies to many or all bootstrap services,
consider adding it to the bootstrap template instead.
-->
<!--- EndBlock(custom) -->

The following sections of CONTRIBUTING.md were generated with
[bootstrap](https://github.com/getoutreach/bootstrap) and are common to all
bootstrap services.

## Dependencies

Make sure you've followed the [Launch Plan](https://outreach-io.atlassian.net/wiki/spaces/EN/pages/695698940/Launch+Plan).
[Set up bootstrap](https://outreach-io.atlassian.net/wiki/spaces/EN/pages/701596137/Services+Checklist) if you're planning on updating bootstrap files.

<!--- Block(devDependencies) -->
<!--- EndBlock(devDependencies) -->

## Building and Testing

<!--- Block(buildCustom) -->
<!--- EndBlock(buildCustom) -->

### Building (Locally)

To produce binaries in the `./bin/` folder, run `make build`.

### Unit Testing

You can run the tests with:

```bash
make test
```
### Integration Testing

Integration tests are tests that require resources such as Kafka or S3.  Please see [Go Testing](https://outreach-io.atlassian.net/wiki/spaces/EN/pages/989693594/Go+Testing) for more details on how to write integration tests.

You can run integration tests with:

```bash
make integration
```

### E2E Testing

E2E tests are tests that require other services, such as `giraffe` or `authz`.  This works with the [Kubernetes dev-environment](https://github.com/getoutreach/dev-environment#getting-started).  If one has not been setup already, this will provision one (which can take a while).

Please see [Go Testing](https://outreach-io.atlassian.net/wiki/spaces/EN/pages/989693594/Go+Testing) for more details on how to write e2e tests.

You can run E2E tests with:

```bash
make e2e
```

## Deploying into the Kubernetes Developer Environment

Create a developer environment using the [Kubernetes dev-environment](https://github.com/getoutreach/dev-environment#getting-started).

Make sure that the service is running locally (run `make devserver`), then swap the provisioned app container with the locally running app inside the cluster:

```bash
devenv local-app vcluster-fs-syncer
```

If you need to update/query the manifests in the dev Kubernetes cluster:

### Create Manifests

```bash
./scripts/deploy-to-dev.sh update
```

### Cleanup Resources

```bash
./scripts/deploy-to-dev.sh delete
```

### Show Manifests

```bash
./scripts/deploy-to-dev.sh show
```


## Releasing

Making releases for this repository follows the process in the [Bootstrap](https://github.com/getoutreach/bootstrap/tree/master/README.md#semver) documentation.
