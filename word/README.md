# word

## How to use

use `make` command

* `make run` : run application in local. I want to use this command to test and debug in local. TODO how to connect local computer by reverse proxy.  
* `make build` : this command is used to build a go execuatable.
* `make docker-build` : to build a docker image.


## How to debug and run in local 

problems : 

* pc connect to consul
* server connect to pc
* about the private address : the other service is on the aws. And their addresses on Consul are private instead of phblic address. 

## TODO Test setting

A whole setting for test mode. 

As production mode. Every thing is private address, no one can connect from local and debug.  Currently, we just have one environment. 

As Dev mode, every thing is using public address. The log, monitor and gateway have a testing instance. 



## TODO 

update template project and slark.