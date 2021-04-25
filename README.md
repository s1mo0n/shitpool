A ProxyPool Without Database
=======

because ip in the pool like a shit, so i don't want a database to save it

### Usage

```bash
# api mode
$ shitpool api
$ shitpool --chk-delay 2 --worker 100 api --addr localhost:3333 --log api.log

# proxy mode
$ shitpool proxy
$ shitpool --chk-delay 2 --worker 100 proxy --httpaddr localhost:4444 --log proxy.log --type all --retry 5

# mix mode
$ shitpool mix
$ shitpool --chk-delay 2 --worker 100 mix --apiaddr localhost:3333 --apilog api.log --httpaddr localhost:4444 --pxylog proxy.log --pxytype all --pxyretry 5

```

### Docker

```bash
$ docker-compose build
$ docker-compose up -d
```

### API
```bash
/count

/get
/get/fuck  # this can bypass the firewall...you know
/get/http
/get/https
/get/socks4
/get/socks5

/get_all
/get_all/fuck
/get_all/http
/get_all/https
/get_all/socks4
/get_all/socks5
```

### Example
```bash
$ shitpool mix &
$ curl -x http://localhost:4444 http://ifconfig.io
xxx.xxx.xxx.xxx
$ curl http://localhost:3333/get
{xxxxxxxxxxxxxxxxxxxxxxxxx}
