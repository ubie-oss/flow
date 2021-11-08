Flow
===============

CL creator for GitOps style CI.

## Overview

![image](./flow.png)

## Usage

To run locally, write your config file and set the path as `FLOW_CONFIG_PATH`.

Then, get an access token from [the GitHub settings page](https://github.com/settings/tokens) and set it as `FLOW_GITHUB_TOKEN`.


```bash
$ export FLOW_CONFIG_PATH=$PWD/config-example.yaml
$ export FLOW_GITHUB_TOKEN=xxxxxx
$ make run
```

Now, the flow app is waiting for pub/sub messages on http://localhost:8080. You can send a dummy request by executing the following command.

```bash
$ make test-message
```

## Test

```bash
$ make test
```

## License

MIT license
