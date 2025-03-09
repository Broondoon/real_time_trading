# real_time_trading

SENG 468 Scalability project repository. Helping you trade stock, reliably.

# Deployment

Requires `docker compose` (`docker-compose` is allegedly outdated) and `golang`.

Ensure you run `docker compose down -v` in between deployment of new Docker instances.

The `-v` flag ensures that the volumes created along with the PostgreSQL DB are also removed.

Deploy with `docker compose up --build`.

# auth-service

The auth service currently employs [gin-gonic](https://gin-gonic.com/docs/introduction/)
package routes under `/authentication/` for `login` and `register` but this may
be replaced with some or all of our interface in `./Shared/`

There is also a `/protected/test` endpoint to validate your JWT token.

# Please Label this

If you need to update the worksite, use go work init ./Shared ./{any other modules we have}

