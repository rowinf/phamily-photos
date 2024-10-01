## Requirements
[go](https://go.dev/)
[postgresql](https://www.postgresql.org/)
[taskfile](https://taskfile.dev/installation)
.env file, see .env.example

create a family record in the database manually, make sure the id is 1.

sign up with a new user using a rest client, see requets.http for example requests

## Running the app

1. `task db:init`
2. `task dev`
3. Open your browser http://localhost:8080/
