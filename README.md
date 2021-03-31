# Export to txtx and FindJSON

`go build -o ipdb2txtx main.go`

```shell
Usage of ./ipdb2txtx:
  -d string
    	test ip address, e.g. ./ipdb2txtx -f ipv4_china.ipdb -d 1.1.1.1
  -f string
    	ipdb file, e.g. ./ipdb2txtx -f ipv4_china.ipdb -l CN >ff.txtx (default "./ipv4_china.ipdb")
  -l string
    	languge (default "EN")
```

```shell
# ./ipdb2txtx -f ipv4_china.ipdb -l CN | more
0.0.0.0	0.255.255.255	保留地址	保留地址	*	*	*	*	*	*	*	*	*	*	*
1.0.0.0	1.0.0.255	CLOUDFLARE.COM	CLOUDFLARE.COM	*	*	*	*	*	*	*	*	*	*	*
1.0.1.0	1.0.3.255	中国	福建	*	*	电信	25.908899	118.125809	Asia/Shanghai	UTC+8	350000	86	CN	AP
```

```shell
# ./ipdb2txtx -d 1.0.2.2 -l CN
BuildTime: 2021-03-24 06:39:20 +0000 UTC
Fields: [country_name region_name city_name owner_domain isp_domain latitude longitude timezone utc_offset china_admin_code idd_code country_code continent_code]

db.Find(1.0.2.2, CN):
[666762 1.0.2.0 1.0.3.255 中国 福建   电信 25.908899 118.125809 Asia/Shanghai UTC+8 350000 86 CN AP] <nil>

db.FindJSON(1.0.2.2, CN):
{"china_admin_code":"350000","city_name":"","continent_code":"AP","country_code":"CN","country_name":"中国","idd_code":"86","ip_end":"1.0.3.255","ip_start":"1.0.2.0","isp_domain":"电信","latitude":"25.908899","longitude":"118.125809","node":"666762","owner_domain":"","region_name":"福建","timezone":"Asia/Shanghai","utc_offset":"UTC+8"} <nil>
```

# ipdb-go

[![TravisCI Build Status](https://travis-ci.org/ipipdotnet/ipdb-go.svg?branch=master)](https://travis-ci.org/ipipdotnet/ipdb-go)
[![Coverage Status](https://coveralls.io/repos/github/ipipdotnet/ipdb-go/badge.svg?branch=master)](https://coveralls.io/github/ipipdotnet/ipdb-go?branch=master)
[![IPDB Database API Document](https://godoc.org/github.com/ipipdotnet/ipdb-go?status.svg)](https://godoc.org/github.com/ipipdotnet/ipdb-go)

IPIP.net officially supported IP database ipdb format parsing library

# Installing
<code>
    go get github.com/ipipdotnet/ipdb-go
</code>

# Code Example

## 支持IPDB格式地级市精度IP离线库(免费版，每周高级版，每日标准版，每日高级版，每日专业版，每日旗舰版)
<pre>
<code>
package main

import (
	"github.com/ipipdotnet/ipdb-go"
	"fmt"
	"log"
)

func main() {
	db, err := ipdb.NewCity("/path/to/city.ipv4.ipdb")
	if err != nil {
		log.Fatal(err)
	}

	db.Reload("/path/to/city.ipv4.ipdb") // 更新 ipdb 文件后可调用 Reload 方法重新加载内容
	
	fmt.Println(db.IsIPv4()) // check database support ip type
	fmt.Println(db.IsIPv6()) // check database support ip type
	fmt.Println(db.BuildTime()) // database build time
	fmt.Println(db.Languages()) // database support language
	fmt.Println(db.Fields()) // database support fields
	
	fmt.Println(db.FindInfo("2001:250:200::", "CN")) // return CityInfo
	fmt.Println(db.Find("1.1.1.1", "CN")) // return []string
	fmt.Println(db.FindMap("118.28.8.8", "CN")) // return map[string]string
	fmt.Println(db.FindInfo("127.0.0.1", "CN")) // return CityInfo
	
	fmt.Println()
}
</code>
</pre>
## 地级市精度库数据字段说明
<pre>
country_name : 国家名字 （每周高级版及其以上版本包含）
region_name  : 省名字   （每周高级版及其以上版本包含）
city_name    : 城市名字 （每周高级版及其以上版本包含）
owner_domain : 所有者   （每周高级版及其以上版本包含）
isp_domain  : 运营商 （每周高级版与每日高级版及其以上版本包含）
latitude  :  纬度   （每日标准版及其以上版本包含）
longitude : 经度    （每日标准版及其以上版本包含）
timezone : 时区     （每日标准版及其以上版本包含）
utc_offset : UTC时区    （每日标准版及其以上版本包含）
china_admin_code : 中国行政区划代码 （每日标准版及其以上版本包含）
idd_code : 国家电话号码前缀 （每日标准版及其以上版本包含）
country_code : 国家2位代码  （每日标准版及其以上版本包含）
continent_code : 大洲代码   （每日标准版及其以上版本包含）
idc : IDC |  VPN   （每日专业版及其以上版本包含）
base_station : 基站 | WIFI （每日专业版及其以上版本包含）
country_code3 : 国家3位代码 （每日专业版及其以上版本包含）
european_union : 是否为欧盟成员国： 1 | 0 （每日专业版及其以上版本包含）
currency_code : 当前国家货币代码    （每日旗舰版及其以上版本包含）
currency_name : 当前国家货币名称    （每日旗舰版及其以上版本包含）
anycast : ANYCAST       （每日旗舰版及其以上版本包含）
</pre>
## 适用于IPDB格式的中国地区 IPv4 区县库
<pre>
db, err := ipdb.NewDistrict("/path/to/quxian.ipdb")
if err != nil {
	log.Fatal(err)
}
fmt.Println(db.IsIPv4())    // check database support ip type
fmt.Println(db.IsIPv6())    // check database support ip type
fmt.Println(db.Languages()) // database support language
fmt.Println(db.Fields())    // database support fields

fmt.Println(db.Find("1.12.7.255", "CN"))
fmt.Println(db.FindMap("2001:250:200::", "CN"))
fmt.Println(db.FindInfo("1.12.7.255", "CN"))

fmt.Println()
</pre>

## 适用于IPDB格式的基站 IPv4 库
<pre>
db, err := ipdb.NewBaseStation("/path/to/station_ip.ipdb")
if err != nil {
	log.Fatal(err)
}

fmt.Println(db.FindMap("223.220.223.255", "CN"))
</pre>