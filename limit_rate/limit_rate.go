package limit_rate

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"time"

	"github.com/arrowltd/slot_backend/env"
	"github.com/mplulu/log"
	"github.com/mplulu/rano"

	"github.com/labstack/echo/v4"
)

const kCheckDuplicateBlockDuration = 5 * time.Minute

type LimitRateCenter struct {
	tlgBot     *rano.Rano
	CounterMap *StringIntMap

	cycleFinished chan bool
	blockedIpMap  *StringIntMap

	blockMutex sync.Mutex
}

func NewCenter(tlgBot *rano.Rano) *LimitRateCenter {
	return &LimitRateCenter{
		tlgBot:        tlgBot,
		cycleFinished: make(chan bool, 1),
		blockedIpMap:  NewStringIntMap(),
	}
}

func (center *LimitRateCenter) Start() {
	if env.E.LimitRate == nil {
		return
	}
	center.startACycle()
	<-center.cycleFinished
	go center.Start()
}

func (center *LimitRateCenter) MiddlewareLimitRate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if env.E.LimitRate == nil {
			return next(c)
		}
		host := c.Request().Host
		path := c.Path()
		ipAddress := c.RealIP()
		if center.isIpAlreadyBlocked(ipAddress) {
			return echo.NewHTTPError(http.StatusTooManyRequests, "")
		}
		center.CounterMap.Increment(fmt.Sprintf("%v%v|%v", host, path, ipAddress), 1)
		return next(c)
	}
}

func (center *LimitRateCenter) startACycle() {
	center.CounterMap = NewStringIntMap()
	go func() {
		<-time.After(1 * time.Second)
		go center.finalizeCycle(center.CounterMap)
		center.cycleFinished <- true
	}()
}

type Request struct {
	URL       string
	IPAddress string
	Count     int
}

func (center *LimitRateCenter) finalizeCycle(counterMap *StringIntMap) {
	requests := []*Request{}
	for key, count := range counterMap.Copy() {
		tokens := strings.Split(key, "|")
		requests = append(requests, &Request{
			URL:       tokens[0],
			IPAddress: tokens[1],
			Count:     count,
		})
	}

	// print top 5
	sort.Slice(requests, func(i int, j int) bool {
		return requests[i].Count > requests[j].Count
	})
	for i := 0; i < 5; i++ {
		if i < len(requests) {
			request := requests[i]
			if env.E.LimitRate.EnableLog {
				log.Log(`LimitRateNo.%v: %v|%v|%v `, i+1, request.URL, request.IPAddress, request.Count)
			}
		}
	}

	total := 0
	willBeBlockedList := []string{}
	for _, request := range requests {
		total += request.Count
		if request.Count > env.E.LimitRate.MaxCount {
			willBeBlockedList = append(willBeBlockedList, request.IPAddress)
		}
	}
	if env.E.LimitRate.EnableLog {
		log.Log("LimitRateTotal: %v", total)
	}
	if total > env.E.LimitRate.MaxTotalCount {
		if len(willBeBlockedList) > 0 {
			go center.blockIps(willBeBlockedList)
		}
	}
}
