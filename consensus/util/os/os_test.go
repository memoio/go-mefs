package os

import (
	"testing"
	"fmt"
)

func Test1(t *testing.T) {
	lp := GetListenPort()
	fmt.Println(GetNumberByList(lp,2))
}
