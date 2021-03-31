package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ipipdotnet/ipdb-go"
)

func main() {
	ipdbFile := flag.String("f", "./ipv4_china.ipdb", "ipdb file, e.g. ./ipdb2txtx -f ipv4_china.ipdb -l CN >ff.txtx")
	ip := flag.String("d", "", "test ip address, e.g. ./ipdb2txtx -f ipv4_china.ipdb -d 1.1.1.1")
	lang := flag.String("l", "EN", "languge")
	flag.Parse()

	// db, err := ipdb.NewCity("../city.free.ipdb")
	db, err := ipdb.NewCity(*ipdbFile)
	if err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	if *ip != "" {
		// test
		fmt.Println()
		fmt.Println("BuildTime:", db.BuildTime())
		fmt.Println("Fields:", db.Fields())

		fmt.Printf("\ndb.Find(%s, %s):\n", *ip, *lang)
		fmt.Println(db.Find(*ip, *lang))

		fmt.Printf("\ndb.FindJSON(%s, %s):\n", *ip, *lang)
		js, err := db.FindJSON(*ip, *lang)
		fmt.Println(string(js), err)
		fmt.Println()
	} else {
		// export
		if err := db.IPv4TXTX(*lang); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func test(db *ipdb.City) {
	fmt.Println(db.FindInfo("2001:250:200::", "CN")) // return CityInfo
	fmt.Println(db.Find("1.1.1.1", "CN"))            // return []string
	fmt.Println(db.FindMap("118.28.8.8", "CN"))      // return map[string]string
	fmt.Println(db.FindInfo("127.0.0.1", "CN"))      // return CityInfo

	fmt.Println(db.Find("1.1.2.2", "CN"))
	fmt.Println(db.Find("1.1.2.2", "EN"))
	fmt.Println(db.Find("1.1.2.2", "XXX"))

	// 注意起止 IP 是指该网段的起止 IP: 1.0.1.0/24, 并非 1.0.1.0 - 1.0.3.255
	fmt.Println(db.Find("1.0.1.1", "CN"))
	// 1.0.2.0/24
	fmt.Println(db.Find("1.0.2.3", "CN"))
}
