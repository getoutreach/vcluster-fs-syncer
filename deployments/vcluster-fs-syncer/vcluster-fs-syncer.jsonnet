// Code managed by Bootstrap, DO NOT MODIFY
// MODIFY THE vcluster-fs-syncer.override.jsonnet INSTEAD
local ok = import 'kubernetes/outreach.libsonnet';
local name = 'vcluster-fs-syncer';
local environment = std.extVar('environment');
local bento = std.extVar('bento');
local cluster = std.extVar('cluster');
local namespace = std.extVar('namespace');
local resources = import './resources.libsonnet';

local isDev = environment == 'development';
local isLocalDev = environment == 'local_development';

local sharedLabels = {
  repo: name,
  bento: bento,
  reporting_team: 'fnd-dt',
};

local all = {
  namespace: ok.Namespace(namespace) {
    metadata+: {
      labels+: sharedLabels,
    },
  },

  // Default configuration for the service, managed by bootstrap.
  // all over configuration should be done in the
  // vcluster-fs-syncer.config.jsonnet file
  configmap: ok.ConfigMap('config', namespace) {
    local this = self,
    // Note: most of the default config is in internal/vcluster-fs-syncer/config.go
    data_:: {},
    data: {
      // We use this.data_ to allow for ez merging in the override.
      'vcluster-fs-syncer.yaml': std.manifestYamlDoc(this.data_),
    },
  },

  deployment: ok.Deployment(name, namespace) {
    metadata+: {
      labels+: sharedLabels,
    },
    spec+: {
      replicas: if (isDev || isLocalDev) then 1 else 2,
      template+: {
        spec+: {
          containers_:: {
            default: ok.Container(name) {
              image: 'gcr.io/outreach-docker/%s:%s' % [name, std.extVar('version')],

              // We don't want to ever pull the same tag multiple times.
              // In dev, this is replaced by sharing docker image cache with Kubernetes
              // so we also don't need to pull images.
              imagePullPolicy: 'IfNotPresent',
              volumeMounts_+:: {
                // default configuration files
                'config-vcluster-fs-syncer': {
                  mountPath: '/run/config/outreach.io/vcluster-fs-syncer.yaml',
                  subPath: 'vcluster-fs-syncer.yaml',
                },
              },
              env_+:: {
                MY_POD_SERVICE_ACCOUNT: ok.FieldRef('spec.serviceAccountName'),
                MY_NAMESPACE: ok.FieldRef('metadata.namespace'),
                MY_POD_NAME: ok.FieldRef('metadata.name'),
                MY_NODE_NAME: ok.FieldRef('spec.nodeName'),
                MY_DEPLOYMENT: name,
                MY_ENVIRONMENT: environment,
                MY_CLUSTER: cluster,
              },
              readinessProbe: {
                httpGet: {
                  path: '/healthz/ready',
                  port: 'http-prom',
                },
                initialDelaySeconds: 5,
                timeoutSeconds: 1,
                periodSeconds: 15,
              },
              livenessProbe: self.readinessProbe {
                initialDelaySeconds: 15,
                httpGet+: {
                  path: '/healthz/live',
                },
              },
              ports_+:: {
                'http-prom': { containerPort: 8000 },
              },
              resources: resources,
            },
          },
          volumes_+:: {
            // default configs
            'config-vcluster-fs-syncer': ok.ConfigMapVolume(ok.ConfigMap('config', namespace)),
          },
        },
      },
    },
  },
};


local override = import './vcluster-fs-syncer.override.jsonnet';
local configuration = import './vcluster-fs-syncer.config.jsonnet';

ok.FilteredList() {
  // Note: configuration overrides the vcluster-fs-syncer.override.jsonnet file,
  // which then overrides the objects found in this file.
  // This is done via a simple key merge, and jsonnet object '+:' notation.
  items_+:: all + (if (isDev || isLocalDev) then {} else {}) + override + configuration,
}
