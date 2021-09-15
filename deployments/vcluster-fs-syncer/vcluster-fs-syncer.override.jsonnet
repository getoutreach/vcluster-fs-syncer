// Put anything here that you want to override (merge) into the original app.jsonnet
local ok = import 'kubernetes/outreach.libsonnet';
local name = 'vcluster-fs-syncer';
local environment = std.extVar('environment');
local bento = std.extVar('bento');
local cluster = std.extVar('cluster');
local namespace = std.extVar('namespace');
local resources = import './resources.libsonnet';
local configuration = import './vcluster-fs-syncer.config.jsonnet';

local isLocalDev = environment == 'local_development';
local isDev = environment == 'development';

// Put custom global variables here
///Block(globalVars)
///EndBlock(globalVars)

local objects = {

  ///Block(override)
  deployment+: {
    kind: 'DaemonSet',
    spec+: {
      replicas: null,
      template+: {
        spec+: {
          serviceAccountName: $.svc_account.metadata.name,
          containers_+:: {
            default+: {
              securityContext: {
                privileged: true,
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
  svc_account: ok.ServiceAccount(name+'-syncer', namespace) {},
  role: ok.ClusterRole(name+'-syncer', namespace) {
    rules: [
      {
        apiGroups: [
          ""
        ],
        resources: [
          "pods"
        ],
        verbs: [
          "get",
          "list",
          "watch"
        ]
      },
    ],
  },
  rolebinding: ok.ClusterRoleBinding(name+"-syncer", namespace) {
    roleRef_:: $.role,
    subjects_:: [$.svc_account],
  },
  ///EndBlock(override)
};

// All objects in this block will only be created in a dev cluster.
// Ideally you'd put VaultSecrets here, or something else.
local dev_objects = {
  ///Block(devoverride)
  ///EndBlock(devoverride)
};

objects + (if (isDev || isLocalDev) then dev_objects else {})
