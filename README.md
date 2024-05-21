# loong
loong is a api gateway. This project is my graduation project. The project now has the following features.
+ Simple HTTP traffic forwarding
+ Simple request validate(Header, JWT)
+ Utilizing pipeline mechanisms to schedule flows 
+ Implement hot reloading
+ IP filtering
+ RateLimit
+ Mocker

Next planned features(after I write my paper or after I finish it):
- [ ] Protocol Transform
- [ ] Support for more filters
- [ ] ...
# usage
If you want to try it, first create the `temp` directory in the project root and create the `pipeline` and `trafficgate` directories.

Build the following file under `traffic`. 
```yml
name: trafficGate-demo
kind: HTTPServer
port: 10080
ipFilter:
  allowIPs:
    - 127.0.0.1
  blockIPs:
    - 1.1.1.1
paths:
  - path: /ping
    backend: pipeline-ping
    methods: 
      - GET
      - POST
      - PUT
    ipFilter:
      allowIPs:
      blockIPs:

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
make
./bin/server
```
if you change your configuration, You just need to run the following command to reload the configuration.
```fish
./bin/client -reload
```

The above two configuration files basically contain the relevant functions of the gateway currently implemented.

# reference projects
+ https://github.com/easegress-io/easegress
+ https://github.com/zehuamama/balancer
+ https://github.com/ermanimer/apigateway



