.DEFAULT_GOAL := build
.PHONY: build

deploy_bucket = is-us-east-1-deployment
aws_profile = default
stack_name = mastostart

build: cli-build lambda-build

cli-build:
	@printf "building $(stack_name) cli:\n"
	@printf "  linux  :: arm64"
	@GOOS=linux GOARCH=arm64 go build -o bin/$(stack_name)-linux-arm64 cmd/$(stack_name)/main.go
	@printf " done.\n"
	@printf "  linux  :: amd64"
	@GOOS=linux GOARCH=amd64 go build -o bin/$(stack_name)-linux-amd64 cmd/$(stack_name)/main.go
	@printf " done.\n"
	@printf "  darwin :: amd64"
	@GOOS=darwin GOARCH=amd64 go build -o bin/$(stack_name)-darwin-amd64 cmd/$(stack_name)/main.go
	@printf " done.\n"
	@printf "  darwin :: arm64"
	@GOOS=darwin GOARCH=arm64 go build -o bin/$(stack_name)-darwin-arm64 cmd/$(stack_name)/main.go
	@printf " done.\n"

tidy:
	@echo "Making mod tidy"
	@go mod tidy

update:
	@echo "Updating $(stack_name)"
	@go get -u ./...
	@go mod tidy

deploy: lambda-build
	@printf "deploying $(stack_name) to aws:\n"
	@mkdir -p build
	@aws --profile $(aws_profile) cloudformation package --template-file aws-cloudformation/template.yaml --s3-bucket $(deploy_bucket) --output-template-file build/out.yaml
	@aws --profile $(aws_profile) cloudformation deploy --template-file build/out.yaml --s3-bucket $(deploy_bucket) --stack-name $(stack_name) --capabilities CAPABILITY_NAMED_IAM
	@printf "done.\n\n"
	@printf "outputs:\n"
	@aws --output json --profile default cloudformation describe-stacks --stack-name mastostart | jq '.Stacks | .[] | .Outputs | reduce .[] as $$i ({}; .[$$i.OutputKey] = $$i.OutputValue)'

lambda-build:
	@printf "building $(stack_name) lambda functions:\n"
	@printf "  mastostart"
	@GOOS=linux GOARCH=arm64 go build -o bin/lambda/mastostart/bootstrap lambda/mastostart/main.go
	@printf " done.\n"

cfdescribe:
	@aws --output json --profile $(aws_profile) cloudformation describe-stack-events --stack-name $(stack_name)

prune:
	@git gc --prune=now
	@git remote prune origin
