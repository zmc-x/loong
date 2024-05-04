# loong
loong is a api gateway. This project is my graduation project. The project now has the following features.
+ Simple HTTP traffic forwarding
+ Simple request validate(Header, JWT)
+ Utilizing pipeline mechanisms to schedule flows 

Next planned features:
- [ ] Distributes Systems
- [ ] Protocol Transform
- [ ] Support for more filters
- [ ] Implementing profile registration
- [ ] Implementing client builds
# usage
If you want to try it, First create the temp directory in the project root and create the `pipeline` and `trafficgate` directories.

Build the following file under `traffic`. (Currently, only the creation of a trafficgate is supported.)
```yml
name: trafficGate-demo
kind: HTTPServer
port: 10080
paths:
  - path: /ping
    backend: pipeline-ping

  - path: /demo
    backend: pipeline-demo
```

Build the following file under `pipeline`.
```yml
name: pipeline-ping
kind: Pipeline
# flow field express handle process for traffic
flow:
  - filter: validator-demo
    jumpIf: 
      invalid: END
  - filter: proxy-demo

filters: 
  - name: validator-demo
    kind: Validator
    headers:
      Content-Type:
        values: 
          - application/json
    jwt:
      algorithm: HS256
      secret: 6d7973656372657

  - name: proxy-demo
    kind: Proxy
    pool: 
      - url: http://127.0.0.1:9096
      - url: http://127.0.0.1:9095
    loadBalance:
      policy: random
```
Then you should build this project in linux. You can run the following command. (Your go version needs to be greater than 1.22)
```fish
go mod tidy
make && ./bin/server
```

# reference projects
+ https://github.com/easegress-io/easegress
+ https://github.com/zehuamama/balancer
+ https://github.com/ermanimer/apigateway



