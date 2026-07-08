NAME := mall-server-docker

BUILD_DIR := build

ENV_FILE=.env


include $(ENV_FILE)

NEXT_VERSION=$(shell echo $(VERSION) | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}')

date=$(shell date +%Y%m%d%H%M%S)

IMAGE_VERSION := v1.4.0

# GOBUILD = env CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -o $(BUILD_DIR)


build:
	docker build -t myapp:$(IMAGE_VERSION) .
	docker login -u=yzyspy@163.com -p YZYspy202502# crpi-7o15kx1cc25d6tg2.cn-hangzhou.personal.cr.aliyuncs.com
	docker tag myapp:$(IMAGE_VERSION) crpi-7o15kx1cc25d6tg2.cn-hangzhou.personal.cr.aliyuncs.com/yangzhongyu/myapp:$(IMAGE_VERSION)
	docker push crpi-7o15kx1cc25d6tg2.cn-hangzhou.personal.cr.aliyuncs.com/yangzhongyu/myapp:$(IMAGE_VERSION)
	@echo "image: registry.cn-hangzhou.aliyuncs.com/yangzhongyu/myapp:$(IMAGE_VERSION)"


generate:
	 protoc --go_out=remote --go-grpc_out=remote  proto/*.proto
	 protoc --go_out=remote --go-grpc_out=remote  proto/rc/decision/*.proto

linux-amd64:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=amd64 GOOS=linux $(GOBUILD)/$@

.PHONY: version
version:
	@echo "Current Version: $(VERSION)"

.PHONY: update-version
update-version:
	@echo "VERSION=$(NEXT_VERSION)" > $(ENV_FILE)
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)


