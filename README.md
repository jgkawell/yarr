# yarr

**yarr** (yet another rss reader) is a web-based feed aggregator which can be used as a personal self-hosted server.

This is a fork of [the original](https://github.com/nkanaev/yarr) which was designed predominantly as a small app to run locally on your Mac or PC. It used an embedded database (sqlite) and had features like menu bar icons. This fork strips all of that away to instead prioritize running the application as a service in orchestration like Kubernetes. The sqlite database has been replaced by postgres and all the features for running locally have been removed. What is left is a super simple microservice with all the same RSS features as the original, fine-tuned to run in orchestration with no local dependencies.

![screenshot](docs/promo.png)

## Usage

### Docker Compose

If you want to just get started right away, simply use the included `docker-compose.yaml` file:

```shell
docker compose up -d
```

### Separate Postgres Instance

If you already have a Postgres instance, you can run the Docker container with any URI you please:

```sh
docker run -it -e YARR_DB=your_postgres_uri_here -p 7070:7070 jgkawell/yarr:latest
```

### Access

Then go to [http://localhost:7070](http://localhost:7070) or [http://server_ip_address:7070](http://server_ip_address:7070) to start using it.

This fork of **yarr** is specifically designed to run as a microservice in orchestration. That could be Docker Compose or Kubernetes and the only requirement is the `YARR_DB` environment variable pointing to an existing Postgres database.

See more:

* [Building from source code](docs/build.md)
* [Fever API support](docs/fever.md)

## credits

[Feather](https://feathericons.com/) for icons.
