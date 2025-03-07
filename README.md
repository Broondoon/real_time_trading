# real_time_trading

SENG 468 Scalability project repository. Helping you trade stock, reliably.

# Deployment

Requires `docker compose` (`docker-compose` is allegedly outdated) and `golang`.

Ensure you run `docker compose down -v` in between deployment of new Docker instances.

The `-v` flag ensures that the volumes created along with the PostgreSQL DB are also removed.

Deploy with `docker compose up --build`.

Alternatively, deployment can be done via Docker Swarm, which is built-in to modern Docker installations.

Deploy a stack to swarm with `docker stack deploy -c <(docker-compose config) real-time-trading`. However, expect this to fail.

First, if you are on linux, run `export $(grep -v '^#' .env | xargs) && docker stack deploy -c docker-stack.yml real-time-trading`. If instead you are on windows, try this code snippet in the Powershell:

<code>(Get-Content .env) | ForEach-Object {
    if ($_ -match "^\s*([^#][^=]+)=(.*)$") {
        Set-Item -Path "Env:$($matches[1])" -Value "$($matches[2])"
    }
}
docker stack deploy -c docker-stack.yml my-stack</code>

The goal is to read the .env file and have your terminal aware of the environment variables, as Docker Swarm does not interpolate .env variables!

This will also likely fail. You need to manually create a network for the stack to use! Try `docker network create --driver overlay --scope swarm go-network`.

Docker Swarm apparently ALSO doesn't support building in the .yaml file. So you're going to have to build the every service via `docker compose build`. A potential source of error will be whether docker-stack.yml has the correct image namges in the image fields.

`DEPRECATED, DO NOT USE`, use the powershell script `env_fix.ps1` on a Windows machine. This file manually edits the docker-compose.yaml with the .env file's contents, as for whatever reason the docker dev team decided that stack doesn't support .env files. Run the script in Powershell via `.\env_fix.ps1 -envFilePath .\.env -dockerComposeFilePath .\docker-compose.yml`. If this fails, make sure Powershell is in Administrator mode and that you have enabled powershell script execution (via `set-executionpolicy remotesigned`).

To increase replicas of a service, use `docker service scale <SERVICE_NAME> = #`.

# auth-service

The auth service currently employs [gin-gonic](https://gin-gonic.com/docs/introduction/)
package routes under `/authentication/` for `login` and `register` but this may
be replaced with some or all of our interface in `./Shared/`

There is also a `/protected/test` endpoint to validate your JWT token.

# Please Label this

If you need to update the worksite, use go work init ./Shared ./{any other modules we have}

