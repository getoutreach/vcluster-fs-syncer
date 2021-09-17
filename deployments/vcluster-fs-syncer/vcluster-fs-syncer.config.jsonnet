// Code managed by Bootstrap - modify only in the blocks
local ok = import 'kubernetes/outreach.libsonnet';
local name = 'vcluster-fs-syncer';
local environment = std.extVar('environment');
local bento = std.extVar('bento');
local cluster = std.extVar('cluster');
local namespace = std.extVar('namespace');
local isDev = environment == 'development';
local isLocalDev = environment == 'local_development';
local devEmail = std.extVar('dev_email');

local custom_attributes = if (isDev || isLocalDev) then { dev_email: devEmail } else {};

// Configuration override for various environments go here.
local configurationOverride = {
  local this = self,
  // Environment level configuration override goes here.
  // Note: `development` and `local_development` refer to different
  // environments. `development` is _inside_ your local k8s cluster
  // while local_development is read by `devconfig.sh`
  environment: {
    local_development: self.development {
      configmap+: {
        data_+:: {
          ListenHost: '127.0.0.1',
          ///Block(localDevelopmentConfig)
          fromPath: './testDir/pods',
          toPath: './testDir/loft/pods',
          ///EndBlock(localDevelopmentConfig)
        },
      },
    },
    development: {
      configmap+: {
        data_+:: {
          ///Block(developmentConfig)
          ///EndBlock(developmentConfig)
        },
      },
    },
    ///Block(environmentConfig)
    ///EndBlock(environmentConfig)
  },

  // Bento level configuration override goes here.
  bento: {
    ///Block(bentoConfig)
    ///EndBlock(bentoConfig)
  },

  // Default configuration for all environments and bentos.
  default: {
    ///Block(defaultConfig)
    configmap+: {
      data_+:: {
        fromPath: '/host_mnt/kubelet/pods',
        toPath: '/host_mnt/loft',
      },
    },
    ///EndBlock(defaultConfig)
  },
};

// configuration merging logic
local env_config = if std.objectHas(configurationOverride.environment, environment) then configurationOverride.environment[environment] else {};
local bento_config = if std.objectHas(configurationOverride.bento, bento) then configurationOverride.bento[bento] else {};

// configuration is the computed value of this service's
// configuration block.
(configurationOverride.default + env_config + bento_config)
