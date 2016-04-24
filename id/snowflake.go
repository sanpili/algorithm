package id

import (
        "time"
        "os"
        "net"
        "math/rand"
        "sync"
)

type Snowflake struct {
        ProcessId uint //进程标识
}

var ipse = getIpSegment()
var serial uint = 0
var lastMillsecond uint64 = 0
var lock  sync.Mutex

func (snow *Snowflake) Next() uint64 {
        now := time.Now()
        millsecond := uint64(now.UnixNano() / 1000000)
        millsecond, serial = getNextSerial(millsecond)
        pid := ipse << 4 | (snow.ProcessId & 0x0000000F) //只取最后4bit
        uid := millsecond << 22 | (uint64(pid) << 10 & 0xfff) | (uint64(serial) & 0x3ff)
        return uid
}

func getNextSerial(millsecond uint64) (uint64, uint) {
        lock.Lock()
        if lastMillsecond < millsecond { //新毫秒
                serial = 0
        }
        if serial >= 0x3ff { //serial用完，等待下一毫秒
                for {
                        <-time.After(time.Microsecond * 500)
                        nsec := uint64(time.Now().UnixNano() / 1000000)
                        if (nsec > millsecond) {
                                millsecond = nsec
                                serial = 0
                                break
                        }
                }
        }
        if serial > 0 {
                serial ++
        } else { //随机开始，id打散
                serial = uint(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(9)) + 1
        }
        lastMillsecond = millsecond
        lock.Unlock()
        return millsecond, serial
}

func getIpSegment() uint {
        addrs, err := net.InterfaceAddrs()
        if err != nil {
                os.Exit(-1)
        }
        for _, addr := range addrs {
                if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                        if ipnet.IP.To4() != nil {
                                return uint(ipnet.IP.To4()[3])
                        }
                }
        }
        os.Exit(-1)
        return 0
}
