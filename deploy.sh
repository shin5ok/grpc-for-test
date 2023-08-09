if test -z $DOMAIN ;
then
    echo "export DOMAIN=<Your domain>"
    echo "  and try it"
    exit
fi

gcloud run deploy \
--source=. \
--region=asia-northeast1 \
--allow-unauthenticated \
--use-http2 \
--set-env-vars=DOMAIN=$DOMAIN,GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT \
grpc-for-test $@
