# real_time_trading

SENG 468 Scalability project repository. Helping you trade stock, reliably.

# Deployment

Requires `docker compose` (`docker-compose` is allegedly outdated) and `golang`.

Ensure you run `docker compose down -v` in between deployment of new Docker instances.

The `-v` flag ensures that the volumes created along with the PostgreSQL DB are also removed.

Deploy with `docker compose up --build`.
