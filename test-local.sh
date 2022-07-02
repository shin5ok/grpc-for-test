grpcurl -plaintext -d '{"id":"v"}' localhost:8080 simple.Simple.GetMessage
grpcurl -plaintext -d '{"message":"foo"}' localhost:8080 simple.Simple.PutMessage
grpcurl -plaintext localhost:8080 simple.Simple.PingPong
