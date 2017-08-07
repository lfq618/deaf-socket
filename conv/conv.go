package conv

import (
	"fmt"
)

//抛出一个类型无法转换的错误
func typeError(val interface{}, t string) error {
	return fmt.Errorf("[%T:%v]无法转换成[%v]类型", val, val, t)
}
