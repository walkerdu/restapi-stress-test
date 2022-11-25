package internal

import (
	"bytes"
	"html/template"
	"io/ioutil"
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
	Sum      int64
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

	urlTemplate  *template.Template
	bodyTemplate *template.Template
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

	// support -d option input data file
	if strings.HasPrefix(mgrIns.Params.Body, "@") {
		dataFile := mgrIns.Params.Body[1:]
		data, err := ioutil.ReadFile(dataFile)
		if err != nil {
			log.Printf("[ERROR] ReadFile:%s err=%s", dataFile, err)
			return
		}

		mgrIns.Params.Body = string(data)
	}
}
func (mgrIns *ClientMgr) Init() error {
	mgrIns.InitParams()
	log.Printf("[DEBGU] StressParams=%+v", mgrIns.Params)

	// url template初始化，添加指定的template function到template中
	urlTemplate, err := template.New("Url").Funcs(funcsMap).Parse(mgrIns.Params.Url)
	if err != nil {
		log.Printf("[ERROR] template New failed, error=%s", err)
		return err
	}
	mgrIns.urlTemplate = urlTemplate

	// body template初始化，添加指定的template function到template中
	bodyTemplate, err := template.New("Body").Funcs(funcsMap).Parse(mgrIns.Params.Body)
	if err != nil {
		log.Printf("[ERROR] template New failed, error=%s", err)
		return err
	}
	mgrIns.bodyTemplate = bodyTemplate

	return nil
}

// 压测结束
func (mgrIns *ClientMgr) IsFinish() bool {
	if mgrIns.Params.Sum <= 0 {
		return false
	}

	if mgrIns.stat.TotalQueryCnts >= mgrIns.Params.Sum {
		log.Printf("[INFO] Finish %d Query", mgrIns.stat.TotalQueryCnts)
		return true
	}

	return false
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

				if mgrIns.IsFinish() {
					break
				}

				var urlBuffer, bodyBuffer bytes.Buffer

				// 生成URL template 请求
				if mgrIns.urlTemplate != nil {
					if err := mgrIns.urlTemplate.Execute(&urlBuffer, nil); err != nil {
						log.Printf("[ERROR] urlTemplate.Execute failed, err=%s", err)
						break
					}
					client.SetUrl(urlBuffer.String())
				}
				// 生成body template 请求
				if mgrIns.bodyTemplate != nil {
					if err := mgrIns.bodyTemplate.Execute(&bodyBuffer, nil); err != nil {
						log.Printf("[ERROR] bodyTemplate.Execute failed, err=%s", err)
						break
					}
					client.SetBody(bodyBuffer.Bytes())
				}

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
