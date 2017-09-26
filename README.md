# mobingi alm-agent

https://learn.mobingi.com/enterprise/api#alm-agent


## install deps

```
make setup
make deps
```

## update depends docker

1. check stable revision of the docker(moby/moby). FYI. https://jenkins.dockerproject.org
2. update `DOCKER_REVISION` file.
3. then update `vendor.conf` below.

```
make vendor.conf
make deps
```
