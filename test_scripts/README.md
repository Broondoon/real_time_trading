Run `source env.sh` to set $HEAP environment variable to a larger variable than
the default. This ensures that jmeter runs with between 4 and 8 gb RAM which is
necessary for the test cases larger than 5k users.

Keep in mind that this is local to your current bash session and is lost if you
close it. Also be sure to clear the Config/stockIds.csv in between calls (as
well as bring down all microservices with `docker compose down -v` so that
the databases are cleared.
