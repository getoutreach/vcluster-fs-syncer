// Code managed by Bootstrap - modify only in the blocks
local ok = import 'kubernetes/outreach.libsonnet';
local name = 'vcluster-fs-syncer';
local environment = std.extVar('environment');
local bento = std.extVar('bento');
local cluster = std.extVar('cluster');
local namespace = std.extVar('namespace');

// Resource override for various enviornments go here.
//
// If a deployment matches on more than one of the overrides then
// the following precedence is observed:
//
// bento > cluster > environment
local resourcesOverride = {
    local this = self,

    // If there is no match for the deployment it will default to
    // the resources defined here.
    default: {
        ///Block(defaultResources)
        requests: {
          cpu: '100m',
          memory: '100Mi'
        },
        limits: self.requests
        ///EndBlock(defaultResources)
    },

    // Environment-level resource overrides go here with the 1st-level keys
    // of the object being environment names.
    environment: {
		local_development: self.development,
        development: {
          requests: {
            cpu: '0',
            memory: '0'
          },
          limits: {}
        },
        ///Block(environmentResources)
        ///EndBlock(environmentResources)
    },

    // Cluster-level resource overrides go here with the 1st-level keys of
    // the object being bento names.
    cluster: {
    	///Block(clusterResources)
    	///EndBlock(clusterResources)
    },

    // Bento-level resource overrides go here with the 1st-level keys of the
    // object being bento names.
    bento: {
        ///Block(bentoResources)
        ///EndBlock(bentoResources)
    }
};

// Resource override merging logic.
local env_resources = if std.objectHas(resourcesOverride.environment, environment) then resourcesOverride.environment[environment] else {};
local cluster_resources = if std.objectHas(resourcesOverride.cluster, cluster) then resourcesOverride.cluster[cluster] else {};
local bento_resources = if std.objectHas(resourcesOverride.bento, bento) then resourcesOverride.bento[bento] else {};

// Computing the final resources object.
(resourcesOverride.default + env_resources + cluster_resources + bento_resources)
