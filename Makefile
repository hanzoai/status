.PHONY: build push deploy deploy-hanzo deploy-lux deploy-pars deploy-zoo deploy-network dev

IMAGE := ghcr.io/hanzoai/status
TAG   := latest
CTX   := do-sfo3-hanzo-k8s

build:
	docker build --no-cache -t $(IMAGE):$(TAG) .

push: build
	docker push $(IMAGE):$(TAG)

# Local dev: run with hanzo brand config
dev:
	docker run --rm -it \
		-p 8080:8080 \
		-v $(PWD)/config/hanzo.yaml:/config/config.yaml:ro \
		$(IMAGE):$(TAG)

dev-lux:
	docker run --rm -it \
		-p 8080:8080 \
		-v $(PWD)/config/lux.yaml:/config/config.yaml:ro \
		$(IMAGE):$(TAG)

dev-pars:
	docker run --rm -it \
		-p 8080:8080 \
		-v $(PWD)/config/pars.yaml:/config/config.yaml:ro \
		$(IMAGE):$(TAG)

dev-zoo:
	docker run --rm -it \
		-p 8080:8080 \
		-v $(PWD)/config/zoo.yaml:/config/config.yaml:ro \
		$(IMAGE):$(TAG)

# Deploy all brands
deploy: deploy-hanzo deploy-lux deploy-pars deploy-zoo deploy-network

# Individual brand deploys
deploy-hanzo:
	kubectl --context $(CTX) apply -k k8s/hanzo/
	kubectl --context $(CTX) -n hanzo rollout restart deployment/status
	kubectl --context $(CTX) -n hanzo rollout status deployment/status --timeout=120s

deploy-lux:
	kubectl --context $(CTX) apply -k k8s/lux/
	kubectl --context $(CTX) -n lux-system rollout restart deployment/status
	kubectl --context $(CTX) -n lux-system rollout status deployment/status --timeout=120s

deploy-pars:
	kubectl --context $(CTX) apply -k k8s/pars/
	kubectl --context $(CTX) -n pars-system rollout restart deployment/status
	kubectl --context $(CTX) -n pars-system rollout status deployment/status --timeout=120s

deploy-zoo:
	kubectl --context $(CTX) apply -k k8s/zoo/
	kubectl --context $(CTX) -n zoo-system rollout restart deployment/status
	kubectl --context $(CTX) -n zoo-system rollout status deployment/status --timeout=120s

deploy-network:
	kubectl --context $(CTX) apply -k k8s/hanzo-network/
	kubectl --context $(CTX) -n hanzo-network rollout restart deployment/status
	kubectl --context $(CTX) -n hanzo-network rollout status deployment/status --timeout=120s

# Update configs only (no image change)
config-hanzo:
	kubectl --context $(CTX) -n hanzo create configmap status-config --from-file=config.yaml=config/hanzo.yaml --dry-run=client -o yaml | kubectl --context $(CTX) apply -f -
	kubectl --context $(CTX) -n hanzo rollout restart deployment/status

config-lux:
	kubectl --context $(CTX) -n lux-system create configmap status-config --from-file=config.yaml=config/lux.yaml --dry-run=client -o yaml | kubectl --context $(CTX) apply -f -
	kubectl --context $(CTX) -n lux-system rollout restart deployment/status

config-pars:
	kubectl --context $(CTX) -n pars-system create configmap status-config --from-file=config.yaml=config/pars.yaml --dry-run=client -o yaml | kubectl --context $(CTX) apply -f -
	kubectl --context $(CTX) -n pars-system rollout restart deployment/status

config-zoo:
	kubectl --context $(CTX) -n zoo-system create configmap status-config --from-file=config.yaml=config/zoo.yaml --dry-run=client -o yaml | kubectl --context $(CTX) apply -f -
	kubectl --context $(CTX) -n zoo-system rollout restart deployment/status

config-network:
	kubectl --context $(CTX) -n hanzo-network create configmap status-config --from-file=config.yaml=config/hanzo-network.yaml --dry-run=client -o yaml | kubectl --context $(CTX) apply -f -
	kubectl --context $(CTX) -n hanzo-network rollout restart deployment/status

# Update all configs
config: config-hanzo config-lux config-pars config-zoo config-network

# Status check
status:
	@echo "=== Hanzo (status.hanzo.ai) ==="
	@kubectl --context $(CTX) -n hanzo get pods -l app=status -o wide 2>/dev/null || echo "  Not deployed"
	@echo ""
	@echo "=== Lux (status.lux.network) ==="
	@kubectl --context $(CTX) -n lux-system get pods -l app=status -o wide 2>/dev/null || echo "  Not deployed"
	@echo ""
	@echo "=== Pars (status.pars.network) ==="
	@kubectl --context $(CTX) -n pars-system get pods -l app=status -o wide 2>/dev/null || echo "  Not deployed"
	@echo ""
	@echo "=== Zoo (status.zoo.network) ==="
	@kubectl --context $(CTX) -n zoo-system get pods -l app=status -o wide 2>/dev/null || echo "  Not deployed"
	@echo ""
	@echo "=== Network (status.hanzo.network) ==="
	@kubectl --context $(CTX) -n hanzo-network get pods -l app=status -o wide 2>/dev/null || echo "  Not deployed"
