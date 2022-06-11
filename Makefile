APP := vcluster-fs-syncer
_ := $(shell ./scripts/bootstrap-lib.sh) 

include .bootstrap/root/Makefile

###Block(targets)
docker-push:
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t "gcr.io/outreach-docker/vcluster-fs-syncer:debug" --push \
		-f deployments/vcluster-fs-syncer/Dockerfile --ssh=default .
###EndBlock(targets)
