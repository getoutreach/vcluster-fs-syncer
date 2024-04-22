local ok = import 'kubernetes/outreach.libsonnet';
local app = (import 'kubernetes/app.libsonnet').info('vcluster-fs-syncer');

local isDev = app.environment == 'development' || app.environment == 'local_development';

local dev_objects = {
  pkgcache: ok.PersistentVolumeClaim('devspace-cache', app.namespace) {
    storage: '10Gi',
  },
  // This volume contains an asdf installation, although only the `install` and
  // `shims` directories.
  asdfinstall: ok.PersistentVolumeClaim('devspace-asdfinstall', app.namespace) {
    storage: '10Gi',
  },
  appcache: ok.PersistentVolumeClaim('devspace-appcache', app.namespace) {
    storage: '2Gi',
  },
};

(if isDev then dev_objects else {})
