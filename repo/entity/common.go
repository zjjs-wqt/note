package entity

import (
	"fmt"
	"time"
)

// Pwd 口令别名用于忽略反序列化
type Pwd string

func (p Pwd) String() string {
	return string(p)
}

var (
	EmptyPwd = []byte(`""`)
)

func (p Pwd) MarshalJSON() ([]byte, error) {
	return EmptyPwd, nil
}

// DateTime 时间格式化
type DateTime time.Time

var cstZone = time.FixedZone("CST", 8*3600) // 东八

func (t DateTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).In(cstZone).Format("2006-01-02 15:04:05"))
	//fmt.Println("stamp = ", stamp)
	return []byte(stamp), nil
}

// DateTimeYMD 时间格式化(仅年月日)
type DateTimeYMD time.Time

func (t DateTimeYMD) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02"))
	//fmt.Println("stamp = ", stamp)
	return []byte(stamp), nil
}
