# restapi-stress-test
stress test for server which based HTTP Rest API

```shell
Usage: %s [options] [URL...]
Options:
	-X, --request <command>
	-u, --user <user:password>
	-H, --header <header,header,header>
	-d, --data <data> or  @data_file
	-s, --sum <the sum of request>
	--qps
```

support go text/template syntax in http request which `url` and `data` field.

```
./restapi-stress-client -u 'test:gogo' -X POST "http://hostname:hostport/xxx/_update/{{uuid}}" -H 'Content-Type: application/json' -d @update.json --qps 1000 -s 100000
./restapi-stress-client -u 'test:gogo' -X POST "http://hostname:hostport/xxx/_update/{{uuid}}" -H 'Content-Type: application/json' --qps 1000 -s 100000 -d '{"params": {
      "publish_time": {{now}},
      "segment": 1,
      "view_count": 0,
      "zone_code": 0,
      "sex": {{rand 1 3}},
      "uid": {{rand 100000000 999999999}},
      "level": {{rand 1 100}},
      "xxx_type": [
        {{rand 1 3}},
        {{rand 3 5}}
      ]
    }
}'
```

support go text/template Functions as follows:
| Function| Description | Usage|
|---|---|---|
| rand | generate [n, m)| {{rand n m}} |
| now | generate unix time (second)| {{now}} |
| uuid | generate uuid string| {{uuid}} |
| date | generate date string, format reference : `time/format.go`| {{date "2006-01-02T15:04:05.999999999Z07:00"}} |
| randDate | like `date`, support random offset second based now| {{date 123 "2006-01-02T15:04:05.999999999Z07:00"}} |
