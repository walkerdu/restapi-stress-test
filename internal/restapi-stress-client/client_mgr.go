package internal

import (
	"log"
	"strings"
	"sync"
	"time"
)

type StressParams struct {
	Method   string
	Url      string
	Body     string
	Headers  map[string]string
	UserName string
	Password string
	Qps      int64
}

type ClientMgr struct {
	// 输入的压测参数
	Params StressParams
	// 临时的Headers，最终会解析入Params中的Headers Map
	MidHeaders string
	// 临时的User:Password，最终会解析入Params中的UserName和Password
	MidUserPass string

	wg   sync.WaitGroup // Wait some task finish
	stop bool
	stat StressStat
}

func NewClientMgr() (*ClientMgr, error) {

	return &ClientMgr{
		Params: StressParams{
			Headers: map[string]string{},
		},
	}, nil
}

func (mgrIns *ClientMgr) InitParams() {
	// http headers
	if len(mgrIns.MidHeaders) != 0 {
		pairsList := strings.Split(mgrIns.MidHeaders, ",")
		for _, headerPair := range pairsList {
			header := strings.Split(headerPair, ":")
			if len(header) != 2 {
				log.Printf("[ERROR] invalid header, header=%s", headerPair)
				continue
			}

			mgrIns.Params.Headers[header[0]] = header[1]
		}
	}

	// http request username and password
	if len(mgrIns.MidUserPass) != 0 {
		userInfo := strings.Split(mgrIns.MidUserPass, ":")
		if len(userInfo) >= 1 {
			mgrIns.Params.UserName = userInfo[0]
		}
		if len(userInfo) >= 2 {
			mgrIns.Params.Password = userInfo[1]
		}
	}
}
func (mgrIns *ClientMgr) Init() {
	mgrIns.InitParams()
	log.Printf("[DEBGU] StressParams=%+v", mgrIns.Params)
}

func (mgrIns *ClientMgr) Run() {
	for loop_i := 1; loop_i <= int(mgrIns.Params.Qps); loop_i++ {
		mgrIns.wg.Add(1)

		go func() {
			client, err := NewClient(mgrIns.Params.Method, mgrIns.Params.Url, mgrIns.Params.Headers, []byte(mgrIns.Params.Body))
			if err != nil {
				log.Printf("[ERROR] NewClient failed, err=%s", err)
				return
			}
			client.SetAuth(mgrIns.Params.UserName, mgrIns.Params.Password)

			defer func() {
				mgrIns.wg.Done()
				client.Destory()
			}()

			for !mgrIns.Stop() {
				beginT := time.Now()

				_, err := client.DoHttp()

				// 统计
				mgrIns.stat.Stat(err == nil, int64(client.Duration))

				duration := time.Now().Sub(beginT)
				if duration >= time.Second {
					continue
				}

				time.Sleep(time.Second - duration)
			}
		}()
	}

	go mgrIns.Stat()

	// wait all goroutine done
	mgrIns.wg.Wait()
}

func (mgrIns *ClientMgr) Stop() bool {
	return mgrIns.stop
}

func (mgrIns *ClientMgr) Stat() {
	for !mgrIns.Stop() {
		log.Printf("[INFO] StressStat Info=%s", mgrIns.stat.String())
		time.Sleep(time.Second)
	}
}
