package snowflake_test

import (
	"go-generator/snowflake"
	"sync"
	"testing"
)


func BenchmarkNextVal(b *testing.B) {
	flake, err := snowflake.NewSnowflake(0, 0)
	if err != nil {
		b.Error(err)
	}
	for i := 0; i < b.N; i++ {
		flake.NextVal()
	}
}

func TestUnique(t *testing.T) {
	var wg sync.WaitGroup
	var check sync.Map
	s, err := snowflake.NewSnowflake(0, 0)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Add(-1)
			val := s.NextVal()
			if duplicate, ok := check.Load(val); ok {
				t.Error(val, duplicate)
				t.Fail()
				return
			}
			check.Store(val, 0)
			if val == 0 {
				t.Fail()
				return
			}
		}()
	}
	wg.Wait()
}
