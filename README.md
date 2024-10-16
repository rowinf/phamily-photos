## Requirements
[go 1.23](https://go.dev/)
[postgresql](https://www.postgresql.org/)
[taskfile](https://taskfile.dev/installation)

## Setup Steps
1. Make sure your postgres db is running on port 5432
2. Create your .env file by starting with a copy of .env.example
3. Update .env and fill in missing values
3. Start the server from a terminal `$ task prod`
4. Open another terminal window
5. Execute command `$ task db:init`
6. Execute command `$ task db:migrate`
7. Execute command `$ task db:seed`
8. Open your browser http://localhost:8080/

## TODO
1. user signup UI
2. family creation UI
3. better login/session security
