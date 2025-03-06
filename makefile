# 定义 run 目标，依赖 build
run: build
	@./app/redis

# 定义 build 目标
build:
	@go build -o app/redis .
	
test:
	@go test ./...
	
