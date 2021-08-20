package common

import "time"

// StartTimer cycle 天之后的 hour 点执行
func StartTimer(f func(), cycle time.Duration, hour int) {
	go func() {
		for {
			now := time.Now()
			next := now.Add(time.Hour * 24 * cycle)
			next = time.Date(next.Year(), next.Month(), next.Day(), hour, 0, 0, 0, next.Location())
			// 测试代码，可以设置几分钟生成一个文件
			// next = time.Date(next.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+hour, 0, 0, next.Location())
			t := time.NewTimer(next.Sub(now))
			<-t.C
			f()
		}
	}()
}
