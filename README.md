# Comments

- I invested around 3h. So I had to make some trade-offs.
- Used Flox/nix for development and environment management.
    - No dockerfile or similar as with flox we can generate OCI images if needed from the environment.
- Basic ci.yml, I disabled linting as fixing lints can take time.
    - I don't push artifacts anywhere, neither deploy.
- TLS/HTTPs locally, I just pushed the certs so you don't have to generate them.
    - Again did not focus on the usual infrastrucutre, multiple certs per environment, pass them as env variable etc.
- Followed the standard go layout: https://go.dev/doc/modules/layout
- Some basic logs but no opentelemtry or metrics collectors:
    - I have a simple dashboard here (4 years old note) https://github.com/vrnvu/microservices-monitoring
- Used net/http for the server.
- Database/schema:
    - I didn't focus a lot on this as I understood from assignment it was not a main focus. 
    - Some projects I've implemented in the space:
        - https://github.com/vrnvu/bitask (KV)
        - https://github.com/vrnvu/rust-open-addresing-linear-probing/tree/master (open addressing, VACCUUM/Vacant ideas on linear data structures, probing strategies)
        - https://github.com/vrnvu/distributed-leveldb (distributed KV in Go with Raft, note 4 years old)
        _ https://github.com/vrnvu/rust-minikeyvalue (distributed KV)
- I didn't use UUID, instead I used nanoids, a really nice alternative.
    - https://adevinta.com/techblog/the-300-bytes-that-saved-millions-optimising-logging-at-scale/
- Not used OpenAPI to generate client/server, API doc.
    - Again understood this was not a main focus or would have been stated directly.
- I did not use a circuit breaker for the DB, again as this was in-memory haven't added them to keep code simple.
- Used promptui for interactie prompt.
    - Not used viper: https://github.com/knadh/koanf?tab=readme-ov-file#alternative-to-viper
    - I followed the recommended Go cli i.e see Golang main contributors:
         - https://github.com/FiloSottile/age/blob/main/cmd/age/age.go
         - Or other offical Google projects how to manage a cli.
- I have some `integration-test.sh`
    - These are "tool agnostic". Using CURL to test the API is a good example. The integration test can then be re-used in other parts of the system. For example we can execute the same integration test in our pre or prod enviornment. 

## Run

```
go run cmd/server/main.go
PORT=9999 go run cmd/server/main.go

go run cmd/cli/main.go --user user quiz
go run cmd/cli/main.go --user user results
go run cmd/cli/main.go --user user statistics
```

## Create a new user and API calls

I did not implement this with the user CLI, simply run

```
curl --cacert localhost.pem https://localhost:8080/health

curl --cacert localhost.pem -X PUT https://localhost:8080/users/newusername
```