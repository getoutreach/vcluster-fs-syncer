# -*- mode: Python -*-

# For more on Extensions, see: https://docs.tilt.dev/extensions.html
load('ext://restart_process', 'docker_build_with_restart')

allow_k8s_contexts('gke_outreach-docker_us-west1_loft-dev-us-west1')

local_resource(
  'compile',
  'make build GOOS=linux GOARCH=amd64 CGO_ENABLED=0',
  deps=['./cmd', './pkg', './internal'],
)

docker_build_with_restart(
  'gcr.io/outreach-docker/vcluster-fs-syncer',
  '.',
  entrypoint=['/app/bin/vcluster-fs-syncer'],
  dockerfile='deployments/vcluster-fs-syncer/Dockerfile.dev',
  only=[
    './bin',
    './deployments/vcluster-fs-syncer',
  ],
  ssh='default',
  live_update=[
    sync('./bin', '/app/bin'),
  ],
)

templated_yaml = local('./scripts/shell-wrapper.sh deploy-to-dev.sh show')
k8s_yaml(templated_yaml)
k8s_resource('vcluster-fs-syncer', port_forwards=8080,
             resource_deps=['compile'])
