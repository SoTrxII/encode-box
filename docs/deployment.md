
A simple but complete deployment of this project could use this format.

First, deploy encode-box and all the backing services using docker-compose.
```yml
version: "3.7"
services:
  # The bot itself, record into raw, unusable files
  pandora:
    image: sotrx/encode-box:0.1.1
    container_name: encode-box
    environment:
      # Pubsub component name
      - PUBSUB_NAME=pubsub
      # Pubsub topic to publish into
      - PUBSUB_TOPIC_PROGRESS=state
      # Object store component name
      - OBJECT_STORE_NAME=object-store
  # Dapr sidecar, defining runtime implementations
  pandora-dapr:
    image: "daprio/daprd:edge"
    command:
      [
        "./daprd",
        "-app-id",
        "encode-box",
        "-dapr-grpc-port",
        "50001",
        "-components-path",
        "/components",
      ]
    # In docker-compose, you have to provide components by sharing a volume
    # this is the dapr/components directory
    volumes:
      - "./components/:/components"
    depends_on:
      - pandora
    network_mode: "service:encode-box"

  # Event broker implementation : redis
  redis:
    image: "redis:alpine"

  # Queue to pull jobs from : rabbitMQ
  rabbit:
    image: "rabbitmq:latest"

  # Object storage implementation : minio
  # /!\ This is just an example, and there is no data persistence /!\
  # For real deployment, see https://docs.min.io/docs/deploy-minio-on-docker-compose.html
  minio:
    image: "minio/minio"
    ports:
      # Minio API
      - "9000:9000"
      # Minio UI
      - "9001:9001"

 ```

Next, define all Dapr components to link every component name used in the env vars to their
implementations
```yml
# components/pubsub.yml
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: pubsub
spec:
  type: pubsub.redis
  version: v1
  metadata:
    - name: redisHost
      value: redis:6379
    - name: redisPassword
      value: ""
```

```yml
# components/object-storage.yml
# Object storage using Minio
# see https://docs.dapr.io/reference/components-reference/supported-bindings/s3/
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: object-store
spec:
  type: bindings.aws.s3
  version: v1
  metadata:
    # Bucket name, should be created BEFOREHAND using the Minio UI
    - name: bucket
      value: recordings-test
    # Anything is fine, its not used in Minio
    - name: region
      value: us-east-1
    # Minio API endpoint
    - name: endpoint
      value: http://minio:9000
    # Mandatory for Minio
    - name: forcePathStyle
      value: true
    # We're using the docker-network without certificates
    - name: disableSSL
      value: true
    # Dapr is encoding all files in B64 before uploading it
    # The following two attributes tells Dapr decode b64 before uploading
    # it on the stoarge backend, and to encode it back when data are retrieved
    - name: encodeBase64
      value: true
    - name: decodeBase64
      value: true
    # An user must be created on Minio using the Minio console to get
    # These attributes
    - name: accessKey
      value: "XnZwvzujlWEzBG5T"
    - name: secretKey
      value: "9p2dKraexj5RzN7kHV9S9H2EAj7RSI9o"
```

Define the message queue, and the subscription to new messages
```yml
# components/message-queue.yml
# https://docs.dapr.io/reference/components-reference/supported-pubsub/setup-rabbitmq/
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: message-queue
spec:
  type: pubsub.rabbitmq
  version: v1
  metadata:
  - name: host
    value: "amqp://message-queue-rabbitmq.tabletop-records.svc.cluster.local:5672"
    # Set this queue to persists when a node is down
  - name: durable
    value: true
    # Delete queue when it has no consumer -> False
  - name: deletedWhenUnused
    value: false
    # Auto ACK a message when a consumer pulls it -> False
  - name: autoAck
    value: false
    # '2' persistant, other : transient (volatile)
  - name: deliveryMode
    value: 2
    # Wether to requeue a failing message
    # In our case, if an encoding failed it will fail every time
    # so don't requeue
  - name: requeueInFailure
    value: false
    # How many message to buffer at a time
  - name: prefetchCount
    value: 1
    # How long to wait before reconnecting
    # on a network failure
  - name: reconnectWait
    value: 0
    # Wether to allow to process multiple
    # message in paralle
  - name: concurrencyMode
    value: parallel
    # Makes the publishers sure that
    # their message was delivered to the broker
  - name: publisherConfirm
    value: false
    # Backoff mode for retry policy
    # 'constant' or 'exponential'
  - name: backOffPolicy
    value: exponential
  - name: backOffInitialInterval
    value: 100
  - name: backOffMaxRetries
    value: 16
    # Send messages that couldn't be
    # processed to another topic
  - name: enableDeadLetter
    value: true
    # Max number of messages in a queue
  - name: maxLen # Optional max message count in a queue
    value: 3000
  - name: exchangeKind
    value: fanout
```

```yml
apiVersion: dapr.io/v1alpha1
kind: Subscription
metadata:
  name: encode
spec:
  topic: encodings
  route: /encode
  pubsubname: message-queue
scopes:
  - encode-box
```