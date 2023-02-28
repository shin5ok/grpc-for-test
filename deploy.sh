gcloud run deploy \
--source=. \
--region=asia-northeast1 \
--allow-unauthenticated \
--use-http2 \
grpc-for-test $@
