# Serverless Roll20 Scrapper

[![codecov](https://codecov.io/gh/SoTrxII/roll20-scrapper/branch/master/graph/badge.svg?token=YI8X1HA6I7)](https://codecov.io/gh/SoTrxII/roll20-scrapper)
[![Docker Image Size](https://badgen.net/docker/size/sotrx/get-players/1.3.0?icon=docker&label=get-players)](https://hub.docker.com/r/sotrx/get-players/)
[![Docker Image Size](https://badgen.net/docker/size/sotrx/join-game/1.3.0?icon=docker&label=join-game)](https://hub.docker.com/r/sotrx/join-game/)
[![Docker Image Size](https://badgen.net/docker/size/sotrx/get-messages/1.3.0?icon=docker&label=get-messages)](https://hub.docker.com/r/sotrx/get-messages/)
[![Docker Image Size](https://badgen.net/docker/size/sotrx/get-summary/1.3.0?icon=docker&label=get-summary)](https://hub.docker.com/r/sotrx/get-summary/)

This project is a serverless (OpenFaas flavored) implementation of a [Roll20](https://roll20.net/welcome) scrapper.
Although all functions share a single core, each of them is distributed as its own container to leverage scalability.

Current functionalities includes :

- Retrieving players from a game
- Retrieving basic infos from a game such a name and image
- Retrieving all messages sent to a chat from a game (including rolls)
- Make the bot account join the game as a player (necessary for other functions)

Full API documentation is available here : https://sotrxii.github.io/roll20-scrapper/

## Configure

This function use a dedicated Roll20 account, as there isn't any Roll20 API. To use it, create a Roll20 account and fill
its info.

Every function needs the following environment variables:

- **ROLL20_USERNAME**: Username of the bot account to use
- **ROLL20_PASSWORD**: Password of the bot account to use
- **ROLL20_BASE_URL**: Roll20 base URL. Value should be "https://app.roll20.net/". This is a variable for future
  proofing and testing purposes.

## Deploying

To deploy the functions, the simplest method is to use [faas-cli](https://docs.openfaas.com/cli/install/).

````shell
# Deploying "get-players"
faas-cli deploy \
 --image "sotrx/get-players:1.3.0"\
 --name "get-players"\
 --gateway <GTW_URL>\
 -e="ROLL20_USERNAME=<BOT_USERNAME>"\
 -e="ROLL20_PASSWORD=<BOT_PASSWORD>"\
 -e="ROLL20_BASE_URL=https://app.roll20.net/"
 
 # Deploying "join-game"
faas-cli deploy \
 --image "sotrx/join-game:1.3.0"\
 --name "join-game"\
 --gateway <GTW_URL>\
 -e="ROLL20_USERNAME=<BOT_USERNAME>"\
 -e="ROLL20_PASSWORD=<BOT_PASSWORD>"\
 -e="ROLL20_BASE_URL=https://app.roll20.net/"
 
  # Deploying "get-messages"
faas-cli deploy \
 --image "sotrx/get-messages:1.3.0"\
 --name "get-messages"\
 --gateway <GTW_URL>\
 -e="ROLL20_USERNAME=<BOT_USERNAME>"\
 -e="ROLL20_PASSWORD=<BOT_PASSWORD>"\
 -e="ROLL20_BASE_URL=https://app.roll20.net/"
 
# Deploying "get-summary"
faas-cli deploy \
 --image "sotrx/get-summary:1.3.0"\
 --name "get-summary"\
 --gateway <GTW_URL>\
 -e="ROLL20_USERNAME=<BOT_USERNAME>"\
 -e="ROLL20_PASSWORD=<BOT_PASSWORD>"\
 -e="ROLL20_BASE_URL=https://app.roll20.net/"
````

### Kubernetes resource

````yaml
# The "function" CRD must be installed on the cluster
# see https://github.com/openfaas/openfaas-operator
apiVersion: openfaas.com/v1
kind: Function
metadata:
  name: join-game
  namespace: openfaas-fn
spec:
  name: join-game
  image: sotrx/join-game:1.3.0
  environment:
    ROLL20_BASE_URL: https://app.roll20.net/
    ROLL20_PASSWORD: <BOT_PASSWORD>
    ROLL20_USERNAME: <BOT_USERNAME>
````

You can also use an [ImageAutomation](https://fluxcd.io/docs/migration/flux-v1-automation-migration/)
from [Flux](https://github.com/fluxcd/flux2) to have a GitOps approach.

## Local development and testing

As they use HTTP trigger only, both the core and the functions can be tested (mocking an incoming HTTP call).

Unit tests are redirecting all Roll20 calls to a mock server, integration tests are using Roll20, and are ignored in
coverage.

If needed, a local [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) based openfaas can be deployed using the
script `setup-dev-env.sh`

The stack YML file (`roll20-scrapper-dev.yml`) is configured to use this local cluster.

### About dependencies

Using a single codebase for multiple functions isn't that simple. From
the [Golang HTTP template page](https://github.com/openfaas/golang-http-template), an import trick is used for the
function to work both in and out of the container. All imports from the `pkg` sub folder must be prefixed
by `handler/function` for the import to work at runtime.

Example:

- Importing the scrapper package -> "handler/function/pkg/scrapper"

Another requirement is to vendor go dependencies. Before building the project, run the following commands

````shell
go mod tidy && go mod vendor
````

