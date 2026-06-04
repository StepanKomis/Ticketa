# Ticketa
Ticketing system built for schools

## Deployment

Ticketa is build from ground up to be as easy to deploy as possible. Anyone from IT student  to skilled admin should have as much control as they need without any ovverhead.

```shell
git clone https://github.com/StepanKomis/Ticketa.git
cd Ticketa
cp .env.example .env
make docker-build
make eploy
```

We use [Make](https://www.gnu.org/software/make/) scripts. Below is table with all the commands with their description.

| command        | description                                                        |
| -------------- | ------------------------------------------------------------------ |
| `build`        | compiles the project into ./build/ticketa                          |
| `run-local`    | deploys postgres instance and starts the binary in ./build/ticketa |
| `docker-build` | builds a ticketa docker image                                      |
| `docker-up`    | deploys docker containers for database and ticketa server itself   |
