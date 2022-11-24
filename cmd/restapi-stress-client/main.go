package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	stressImp "github.com/walkerdu/restapi-stress-test/internal/restapi-stress-client"
)

var (
	usage = `Usage: %s [options] [URL...]
Options:
	-X, --request <command>
	-u, --user <user:password>
	-H, --header <header,header,header>
	-d, --data <data> or  @data_file
	--qps

`
	Usage = func() {
		//fmt.Println(fmt.Sprintf("Usage of %s:\n", os.Args[0]))
		fmt.Printf(usage, os.Args[0])
	}
)

func main() {
	flag.Usage = Usage
	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(1)
	}

	clientMgr, err := stressImp.NewClientMgr()
	if err != nil {
		log.Printf("[ERROR] NewClientMgr failed, err=%s", err)
		return
	}

	flag.StringVar(&clientMgr.Params.Method, "X", "GET", "http request method")
	flag.StringVar(&clientMgr.Params.Method, "request", "GET", "http request method")
	flag.StringVar(&clientMgr.MidUserPass, "u", "", "http request user and password")
	flag.StringVar(&clientMgr.MidUserPass, "user", "", "http request user and password")
	flag.StringVar(&clientMgr.MidHeaders, "H", "", "http request headers")
	flag.StringVar(&clientMgr.MidHeaders, "header", "", "http request headers")
	flag.StringVar(&clientMgr.Params.Body, "d", "", "http request data")
	flag.StringVar(&clientMgr.Params.Body, "data", "", "http request data")
	flag.Int64Var(&clientMgr.Params.Qps, "qps", 1, "http request QPS")

	flag.Parse()

	// 输入参数中没有options的默认位URL参数，可以在任意位置
	for flag.NArg() > 0 {
		if len(clientMgr.Params.Url) == 0 {
			clientMgr.Params.Url = flag.Args()[0]
		}
		os.Args = flag.Args()[0:]
		flag.Parse()
	}

	clientMgr.Init()
	clientMgr.Run()
}
