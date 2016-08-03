.DEFAULT: all
.PHONY: all clean image publish-image minikube-pubish kube-deploy kube-redeploy kube-unedeploy

all: image

clean:
	go clean
	rm -rf ./build

godeps=$(shell go get $1 && go list -f '{{join .Deps "\n"}}' $1 | grep -v /vendor/ | xargs go list -f '{{if not .Standard}}{{ $$dep := . }}{{range .GoFiles}}{{$$dep.Dir}}/{{.}} {{end}}{{end}}')

DEPS=$(call godeps,./cmd/weave-npc)

cmd/weave-npc/weave-npc: $(DEPS)
cmd/weave-npc/weave-npc: cmd/weave-npc/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ cmd/weave-npc/main.go

build/.image.done: cmd/weave-npc/Dockerfile cmd/weave-npc/weave-npc
	mkdir -p build
	cp $^ build
	sudo docker build -t weaveworks/weave-npc -f build/Dockerfile ./build
	touch $@

image: build/.image.done

publish-image: image
	sudo docker push weaveworks/weave-npc

minikube-publish: image
	sudo docker save weaveworks/weave-npc | (eval $$(minikube docker-env) && docker load)

kube-deploy: all
	kubectl create -f k8s/daemonset.yaml

kube-redeploy: all
	kubectl delete pods --namespace kube-system -l k8s-app=weave-npc

kube-undeploy:
	kubectl delete -f k8s/daemonset.yaml
