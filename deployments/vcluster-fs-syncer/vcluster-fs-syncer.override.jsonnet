// Copyright 2023 Outreach Corporation. All Rights Reserved.
//
// Description: This file is automatically merged into the 'vcluster-fs-syncer.jsonnet' file.
// Configuration should go into the 'vcluster-fs-syncer.config.jsonnet' file, or in the relevant
// file in the configs/ directory.
//
// Managed: true

local ok = import 'kubernetes/outreach.libsonnet';
local app = (import 'kubernetes/app.libsonnet').info('vcluster-fs-syncer');
local isDev = (app.environment == 'development' || app.environment == 'local_development');

// Put custom global variables here
// <<Stencil::Block(globalVars)>>

// <</Stencil::Block>>

// Objects contains kubernetes objects (or resources) that should be created in
// all environments.
// Note: If creating an HPA, you will need to remove the deployment.replica so it does not conflict.
// Ex: deployment+: {spec+: { replicas: null, }, },
local objects = {
  // <<Stencil::Block(override)>>
  deployment+: {
    kind: 'DaemonSet',
    spec+: {
      replicas: null,
      strategy: null,
      template+: {
        spec+: {
          serviceAccountName: $.svc_account.metadata.name,
          containers_+:: {
            default+: {
              securityContext: {
                // Required for Bidirectional mount propagation
                privileged: true,
                capabilities: {
                  add: ['SYS_ADMIN'],
                },
                allowPrivilegeEscalation: true,
              },
              volumeMounts_+:: {
                'var-lib': {
                  mountPath: '/host_mnt',

                  // Enables our bind mounts from syncer to be reflected
                  // on the host
                  mountPropagation: 'Bidirectional',
                },
              },
            },
          },
          volumes_+:: {
            'var-lib': {
              hostPath: {
                path: '/var/lib',
                type: 'Directory',
              },
            },
          },
        },
      },
    },
  },
  svc_account: ok.ServiceAccount(app.name + '-syncer', app.namespace) {},
  role: ok.ClusterRole(app.name + '-syncer', app.namespace) {
    rules: [
      {
        apiGroups: [
          '',
        ],
        resources: [
          'pods',
        ],
        verbs: [
          'get',
          'list',
          'watch',
        ],
      },
    ],
  },
  rolebinding: ok.ClusterRoleBinding(app.name + '-syncer', app.namespace) {
    roleRef_:: $.role,
    subjects_:: [$.svc_account],
  },
  // <</Stencil::Block>>
};

// dev_objects contains kubernetes objects (or resources) that should be created
// ONLY in a development environment. This includes the E2E environment.
local dev_objects = {
  // <<Stencil::Block(devoverride)>>

  // <</Stencil::Block>>
};

// overrideMixins contains a list of files to include as mixins into
// the override file.
local overrideMixins = [
  // <<Stencil::Block(overrideMixins)>>

  // <</Stencil::Block>>
];

local mergedOverrideMixins = std.foldl(function(x, y) (x + y), overrideMixins, {});
mergedOverrideMixins + objects + (if isDev then dev_objects else {})
