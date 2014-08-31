# groupcache-example
The project is built on top of of [groupcache-db-experiment](https://github.com/capotej/groupcache-db-experiment) with etcd[https://github.com/coreos/etcd] support.

This project simulates a scenario wherein a few frontends running [groupcache](http://github.com/golang/groupcache) are fronting a slow database. For more detail,  See capotej [blog post](http://www.capotej.com/blog/2013/07/28/playing-with-groupcache/) about it for more details.

# Getting it running
The following commands will set up this topology:
![groupcache topology](https://raw.github.com/capotej/groupcache-db-experiment/master/topology.png)

### Build everything

1. ```git clone git@github.com:senarukana/groupcache-example```
2. ```sh build.sh```

### Start DB server

1. ```cd dbserver && ./dbserver```

This starts a delibrately slow k/v datastore on :8000

### Start Multiple Frontends

1. ```cd cacheserver```
2. ```./cacheserver -port 8001```
3. ```./cacheserver -port 8002```
4. ```./cacheserver -port 8003```

### Use the CLI to set/get values

SET; Set the data to db
CGET: Get the data from the groupcache
GET: Get the data from db

1. ```cd cli```
2. ```./cli set foo bar``
3. ```./cli get foo``` should see bar in 300 ms
4. ```./cli cget foo``` should see bar in 300ms (cache is loaded)
5. ```./cli cget foo``` should see bar instantly

### More Scenario
See capotej [blog post](http://www.capotej.com/blog/2013/07/28/playing-with-groupcache/) for more details.
