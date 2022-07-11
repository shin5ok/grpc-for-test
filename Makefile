
go:
	protoc -Iproto --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false proto/simple.proto

python:
	pip install -r requirements.txt
	python -m grpc_tools.protoc -Iproto --python_out=pb --grpc_python_out=pb proto/simple.proto