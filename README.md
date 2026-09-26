# Overview
The registry-exporter deployment queries a local docker registry. It gathers statistics for each image it gathers statistics like image size, total size, number of tags, architecture and os. It works with any docker v2 compatible registry that offers a _catalog api, this includes:
- CNCF distribution (the v2 docker registry)
- Harbor
- Sonatype Nexus
- Gitea

Most public registries have disabled the _catalog api.
With the grafana dashboard it is easy to see which images consumes the most space for example.

# Environment variables
The following enviroment variables are supported in the values.yaml:
| Variable | Description |
| -------- | -------- |
| REFRESH_SECONDS | The amount of seconds between successive polls of the registry |
| PORT_NUMBER | The port that is used for publishing the metrics |
| REGISTRY | The domainname for the local registry to query |
| SCHEME | the scheme to contact the registry, one of http or https |


# Grafana screenshot
![Grafana](https://raw.githubusercontent.com/gounix/registry-exporter/main/registry-exporter.png)

# Container
[docker hub](https://hub.docker.com/r/gounix/registry-exporter)

# Sources
[GitHub](https://github.com/gounix/registry-exporter/src)
