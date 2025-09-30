# Kafka Connect

This Chart will install the Confluent's Kafka connect, optionally creating
a Kafka too.

## Installing plugin

To add your plugin to the running pod, pass the following content to your
installation:

```yaml
extraVolumeMounts:
  - name: plugin
    mountPath: /usr/share/confluent-hub-components
extraVolumes:
  - name: plugin
    emptyDir: {}
initContainers:
  - name: init-plugin
    image: confluentinc/cp-kafka-connect:7.2.2
    command:
      - sh
      - -c
      - confluent-hub install mongodb/kafka-connect-mongodb:1.8.0 --no-prompt
    volumeMounts:
      - name: plugin
        mountPath: /usr/share/confluent-hub-components
```

## Bring your own Kafka

To specify another Kafka as a broker, simply pass the following content to
your installation:

```yaml
kafka:
  create: false
configMapPairs:
  CONNECT_BOOTSTRAP_SERVERS: kafka:9092
```

## Bring your own Schema Registry

To specify another Kafka as a broker, simply pass the following content to
your installation:

```yaml
schema-registry:
  create: false
configMapPairs:
  CONNECT_VALUE_CONVERTER_SCHEMA_REGISTRY_URL: http://schema-registry:8081
```

## Schema Registry Config

If you install the Chart using a name other than the Chart's name, be sure to
modify the following value accordingly:

```yaml
schema-registry:
  externalKafka:
    brokers:
      - PLAINTEXT://kafka-connect-kafka:9092
```
