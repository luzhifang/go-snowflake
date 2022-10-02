package snowflake

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	epoch          = int64(1640966400000)              // 纪元：2022-01-01 00:00:00，有效期69年
	timestampBits  = uint(41)                          // 时间戳占用位数
	serverIdBits   = uint(10)                          // 机器ID所占位数
	sequenceBits   = uint(12)                          // 序列所占的位数
	timestampMax   = int64(-1 ^ (-1 << timestampBits)) // 时间戳最大值
	serverIdMax    = int64(-1 ^ (-1 << serverIdBits))  // 支持的最大机器ID数量
	sequenceMask   = int64(-1 ^ (-1 << sequenceBits))  // 支持的最大序列ID数量
	serverIdShift  = sequenceBits                      // 机器ID左移位数
	timestampShift = sequenceBits + serverIdBits       // 时间戳左移位数
)

type Snowflake struct {
	sync.Mutex
	timestamp int64
	serverId  int64
	sequence  int64
}

func NewSnowflake(serverId int64) (*Snowflake, error) {
	// serverId 只在程序启动时进行初始化
	// serverId 可以用hash(本机IP或者主机名)
	// 判断重复，可以用分布式缓存加分布式锁
	// 如果重复，可以继续hash，直到不重复为止
	if serverId > serverIdMax {
		return nil, fmt.Errorf("serverId must be between 0 and %d", serverIdMax)
	}
	return &Snowflake{
		timestamp: 0,
		serverId:  serverId,
		sequence:  0,
	}, nil
}

func (s *Snowflake) NextId() (r int64, err error) {
	s.Lock()
	now := time.Now().UnixNano() / 1e6
	if s.timestamp < now {
		// 当前时间增加了，序列号清0
		s.sequence = 0
		s.timestamp = now
	} else if s.timestamp == now {
		// 还处在上一毫秒，序列号+1
		// 还得判断序列号增到最大值之后，序列号会清0，此时时间戳需要+1，不然会出现重复
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			s.timestamp++
		}
	} else {
		// 时间回拨，序列号清0，时间戳+1
		s.timestamp++
		s.sequence = 0
	}
	t := s.timestamp - epoch
	if t > timestampMax {
		s.Unlock()
		log.Printf("epoch must be between 0 and %d\n", timestampMax)
		return r, fmt.Errorf("epoch must be between 0 and %d", timestampMax)
	}
	r = (t << timestampShift) | (s.serverId << serverIdShift) | s.sequence
	s.Unlock()
	return r, nil
}
