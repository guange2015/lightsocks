package local

import (
	"fmt"
	"testing"
)

func Test_Local(t *testing.T) {

	buf1 := make([]byte, 10)
	for i := 0; i < 10; i++ {
		buf1[i] = byte(i)
	}

	buf2 := make([]byte, len(buf1)+1)
	buf2[0] = 11
	copy(buf2[1:], buf1)

	t.Log(fmt.Printf("%v, %v", buf1, buf2))
}
