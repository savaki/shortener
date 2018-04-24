shortener
---------

URL shortener that uses API Gateway

### Prerequisites

To use this project, you'll need to:

* ensure aws credentials are configured in environment `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
* ensure `up` is installed.  [Download Up](https://up.docs.apex.sh/)

This particular project makes use of go, so golang would need to be installed.  However, up
supports most modern languages including python.  Go was just for time/performance reasons.

### Deployment

To deploy the project, type the following:

```bash
up
```

The first time this is called, `up` will create an api gateway service with the name `shortener` 
(defined in up.json) along with an iam role for the service.

Subsequent deploys will simply update the service.

### Usage with environments

To specify the environment to deploy to, you can use:

```bash
up deploy staging    # deploys to staging
up deploy production # deploys to production
``` 

`up` supports stage specific configurations that can be placed in `up.json`


### Try it out

[https://6vrp7g0fqf.execute-api.us-east-1.amazonaws.com/staging/](https://6vrp7g0fqf.execute-api.us-east-1.amazonaws.com/staging/})

#### Register a new shortened url

```bash
curl -v -X POST -d key=def -d url=http://google.com https://6vrp7g0fqf.execute-api.us-east-1.amazonaws.com/staging/
```

### Make use of the shortened url

```bash
curl -L https://6vrp7g0fqf.execute-api.us-east-1.amazonaws.com/staging/def
```

### Batteries Not Included

This POC contains a number of interesting features:

* Uses cloudfront URL to reduce S3 costs and improve latency
* Internal LRU again speeds up response times

However, there are quite a few things to do to make the service production-ready.

* authentication to /register endpoint
* use signed cloudfront urls (to avoid exposing cloudfront publicly)
* validate inputs
