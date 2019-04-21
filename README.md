A gateway for importing traces from Zipkin or Jaeger instrumented clients into
ElasticSearch APM.

It exposes two API endpoints:

* A Zipkin-compatible API endpoint at `http://0.0.0.0:9411/api/v2/spans`.
* A Jaeger Collector-compatible API endpoint at `http://0.0.0.0:14268/api/traces`.

These API endpoints can be used to import data into ElasticSearch using the APM server.

# Runing

Given an APM server address, run the Docker image:

```
docker run -p 9411:9411 -p 14268:14268 -e APM_ENDPOINT=http://127.0.0.1:8200/intake/v2/events justinbarrick/apm-gateway:dev
```

Traces can then be sent to the Jaeger or Zipkin endpoints and viewed in APM.

# Deploying

There are very basic demo manifests in `deploy/` that can be used to deploy an
ElasticSearch, Kibana, APM server, and APM gateway stack.

```
kubectl apply -f deploy/
```
