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
    },
  },
  pdb: null,
  ///EndBlock(override)
};

// All objects in this block will only be created in a dev cluster.
// Ideally you'd put VaultSecrets here, or something else.
local dev_objects = {
  service+: {
    metadata+: {
      annotations+: {
        // Allow everyone AdminGW gRPCUI access in dev environment
        'outreach.io/admingw-allow-grpc-1000000': '.* Everyone',
      },
    },
  },

  ///Block(devoverride)
  ///EndBlock(devoverride)
};

objects + (if (isDev || isLocalDev) then dev_objects else {})
