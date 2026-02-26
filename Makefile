.PHONY: build push deploy-hanzo deploy-lux dev

IMAGE := ghcr.io/hanzoai/status
TAG   := latest

build:
	docker build -t $(IMAGE):$(TAG) .

push: build
	docker push $(IMAGE):$(TAG)

# Local dev: run with hanzo brand config
dev:
	docker run --rm -it \
		-p 8080:8080 \
		-v $(PWD)/config/hanzo.yaml:/config/config.yaml:ro \
		$(IMAGE):$(TAG)

# Deploy to hanzo-k8s
deploy-hanzo:
	kubectl --context hanzo-k8s apply -k k8s/hanzo/

deploy-lux:
	kubectl --context hanzo-k8s apply -k k8s/lux/
