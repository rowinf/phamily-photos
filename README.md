## Requirements
- [go](https://go.dev/)
- [postgresql](https://www.postgresql.org/)
- [taskfile](https://taskfile.dev/installation)
- .env file, see .env.example

## Running the app

### Setup
1. create a family record in the database manually, make sure the id is 1.
2. sign up with a new user using a rest client, see requets.http for example requests

### Commands
1. `task db:init`
2. `task dev`
3. Open your browser http://localhost:8080/

## TODO
1. user signup UI
2. family creation UI
3. better login/session security
