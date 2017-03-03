CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o go-newrelic-plugin -v
