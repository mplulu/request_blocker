package limit_rate

import (
	"fmt"
	"os/exec"

	"github.com/mplulu/log"

	"time"
)

type BlockedIP struct {
	ip        string
	blockedAt time.Time
}

func (center *LimitRateCenter) isIpAlreadyBlocked(ip string) bool {
	return center.blockedIpMap.Has(ip)
}

func (center *LimitRateCenter) blockIps(inputIps []string) {
	center.blockMutex.Lock()
	defer center.blockMutex.Unlock()
	ips := []string{}
	for _, inputIP := range inputIps {
		if !center.isIpAlreadyBlocked(inputIP) {
			ips = append(ips, inputIP)
		}
	}
	log.Log("will block %v", ips)
	for _, ip := range ips {
		center.blockedIpMap.Set(ip, 1)
	}
	now := time.Now()
	for _, ip := range ips {
		cmd := exec.Command("sh", "-c", fmt.Sprintf(
			`echo 'deny %v; ## %v' | sudo tee -a /etc/nginx/conf.d/blacklist_ips.conf`,
			ip, now.String()))
		stdout, err := cmd.CombinedOutput()
		if err != nil {
			log.LogSerious("LimitRate blockIpsErr %v %v", string(stdout), err)
			return
		}
	}
	cmd := exec.Command("sh", "-c", `sudo nginx -t && sudo nginx -s reload`)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		log.LogSerious("LimitRate restart nginx %v %v", string(stdout), err)
	}
	message := log.Log("done block %v", ips)
	center.tlgBot.Send(message)
}
