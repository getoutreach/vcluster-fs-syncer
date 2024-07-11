APP := vcluster-fs-syncer
OSS := true
_ := $(shell ./scripts/devbase.sh)

include .bootstrap/root/Makefile

## <<Stencil::Block(targets)>>
docker-push:
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t "gcr.io/outreach-docker/vcluster-fs-syncer:debug" --push \
		-f deployments/vcluster-fs-syncer/Dockerfile --ssh=default .

post-stencil::
	./scripts/shell-wrapper.sh catalog-sync.sh
	make fmt
	yarn upgrade
## <</Stencil::Block>>
