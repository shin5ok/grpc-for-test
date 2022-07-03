# for Cloud Run
# Part_of_Cloud_Run_Host_exclude_port
CLOUD_RUN_HOST=$1
grpcurl -vv -proto proto/simple.proto -d '{"id":"test"}' ${CLOUD_RUN_HOST}:443 simple.Simple.GetMessage
sleep 1
grpcurl -vv -proto proto/simple.proto -d '{"message":"test message"}' ${CLOUD_RUN_HOST}:443 simple.Simple.PutMessage

# for Cloud Run service required auth
TOKEN=$(gcloud auth print-identity-token)
grpcurl -H "Authorization: Bearer $TOKEN" -proto proto/simple.proto ${CLOUD_RUN_HOST}:443 simple.Simple.PingPong