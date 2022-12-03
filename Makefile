BASE_PATH := $(shell pwd)
CMD_PATH := "cmd/"
CMD_SRCS := $(shell cd $(CMD_PATH); ls)
GO_COMPILE:=GOOS=linux GOARCH=amd64 go build 

.SILENT:
init:
	echo "ℹ️  INITIALIZING PROJECT..."
	npm install
	go mod tidy

ifdef profile
	cdk bootstrap --profile ${profile}
else
	cdk bootstrap
endif

.PHONY: clean compile_all
clean:
	echo "ℹ️  CLEANING ALL  BUILD FILES..."
	for src in $(CMD_SRCS); do \
		echo "- $$src"; \
		cd $(BASE_PATH)/$(CMD_PATH)$$src && \
		if [ -f $$src ]; then \
			rm $$src; \
		fi; \
	done
	echo "\n"

compile_all: clean
	echo "ℹ️  STARTING TO COMPILE ALL..."
	for src in $(CMD_SRCS); do \
		echo "- $$src"; \
		cd $(BASE_PATH)/$(CMD_PATH)$$src && \
		$(GO_COMPILE) -o $$src main.go; \
	done
	echo "\n"; \

deploy: compile_all
	echo "🚀 Deploying stack..."

ifdef profile
	cdk deploy --profile ${profile}
else
	cdk deploy
endif