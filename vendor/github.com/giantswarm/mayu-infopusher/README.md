[![CircleCI](https://circleci.com/gh/giantswarm/mayu-infopusher.svg?&style=shield&circle-token=b8f4fbcc688135bdb5e8466cdfec46c42ddd5ce3)](https://circleci.com/gh/giantswarm/mayu-infopusher)
[![Docker Repository on Quay](https://quay.io/repository/giantswarm/mayu-infopusher/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/mayu-infopusher)

# infopusher
Tool that is originally included in https://github.com/giantswarm/mayu


It is used for gathering info about machine and requesting cloudconfig/iginition for this machine from `mayu`.

# Docker image
Image is avaiable at `quay.io` - https://quay.io/repository/giantswarm/mayu-infopusher

When runing in docker container, you have to provide few extra opts (because `infopusher` needs some hardware information and also access to ipmi dev)
```
docker run --net=host --privileged=true -v /sys:/sys -v /dev:/dev -it quay.io/giantswarm/mayu-infopusher:33f832fcf8c0fa3a80300274ebb801ec2a87d3e3
```

