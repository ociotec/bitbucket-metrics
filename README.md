# bitbucket-metrics

Service to expose Bitbucket metrics written in Go

## Configuration

Required environment variables:

* `BASE_URL` Bitbucket base URL to access (do not include on this variable the API sub URI).
* `USERNAME` Username to authenticate with.
* `PASSWORD` Password to authenticate with.

Optional environment variables:

* `CONFIG` Configuration file to be used (it's decribed later).
* `LOG_LEVEL` Log level to be used: `debug`, `info` (default one), `warn`, `error`, `fatal`, `panic`.

Rest of values should be configured in a YAML file, use `config.example.yaml` as en example one.

```yaml
bitbucket:
  api_page_size: 100
  metrics:
    hostname: localhost
    port: 8080
    path: /metrics
    period_in_seconds: 3600
  projects:
    include:
      - project1
      - project2
```

## Metrics

Additionally to go metrics, these are the exposed metrics:

* `bitbucket_projects`
* `bitbucket_repositories`
* `bitbucket_prs_by_author` labeled by `project`, `repo` & `author`
* `bitbucket_prs_by_reviewer` labeled by `project`, `repo` & `reviewer`
* `bitbucket_collect_time` last metrics collection time in milliseconds

## Docker

To run it with Docker:

```bash
# --rm Removes the container once it stops
# -it Run it interactively
# -v Mount a volume with the config.yaml file
# -v Mount SSL certificates from Docker host to the container
# -e Several environment variables
docker run --rm \
           -it \
           -v $(pwd)/config.yaml:/config.yaml \
           -v /etc/ssl/certs:/etc/ssl/certs \
           -e BASE_URL=https://bitbucket-url \
           -e USERNAME=the-username \
           -e PASSWORD=the-password \
           ociotec/bitbucket-metrics:latest
```

## License

Copyright 2025 Emilio González Montaña

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License
for the specific language governing permissions and limitations under the License.

Commons Clause Restriction

The Software is provided under the Apache License, Version 2.0, as modified by
the Commons Clause below:

The Commons Clause is an addendum to the Apache License that restricts the ability
to sell the Software.

The license granted herein is expressly made subject to the following Commons
Clause condition:

"Notwithstanding any other provision of the License, the license granted herein
does not include the right to Sell the Software. For purposes of the foregoing,
'Sell' means practicing any or all of the rights granted to you under the License
to provide to third parties, for a fee or other consideration (including without
limitation fees for hosting or consulting/support services related to the Software),
a product or service whose value derives, entirely or substantially, from the
functionality of the Software. Any license notice or attribution required by the
License must also include this Commons Clause."

For commercial licensing or modification permissions, please contact:
egmontana@hotmail.com
