# Giantswarm Custom Metric Autoscaler

This is a basic Kubernetes Cronjob that scrapes a prometheus metric indicating required number of nodes for the cluster
and reconciles this against the [Giantswarm](https://giantswarm.io) API.

## Installation

```console
$ helm install contrib/g8s-scaler
```

## Configuration

| Parameter                    | Description                                     | Default                        |
| ---------------------------- | ----------------------------------------------- | ------------------------------ |
| `scalerEnabled`              | to enable or disable the scaler                 | false                          |
| `scaler.frequency`           | frequeny (cron syntax) the scaler should scrape | "_/10 _ \* \* \*"              |
| `scaler.image`               | docker image                                    | quay.io/pickledrick/g8s-scaler |
| `scaler.tag`                 | docker image tag                                | latest                         |
| `scaler.config.token`        | Giantswarm API token                            | token                          |
| `scaler.config.cluster`      | Giantswarm API unique cluster ID                | cluster                        |
| `scaler.config.fetch_url`    | URL to fetch metrics                            | http://url                     |
| `scaler.config.fetch_path`   | Path metrics are exposed                        | /metrics                       |
| `scaler.config.fetch_metric` | Name of metric exposed                          | required_nodes                 |
| `scaler.config.g8s_api`      | Giantswarm API endpoint                         | https://api/v4/clusters/       |
