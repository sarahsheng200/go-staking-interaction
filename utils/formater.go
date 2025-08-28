package utils

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

func IsEmptyOrSpaceString(s string) bool {
	// 去除所有空白字符后判断是否为空
	return strings.TrimSpace(s) == ""
}

func StringToBigInt(s string) (*big.Int, error) {
	res := new(big.Int)
	if _, ok := res.SetString(s, 10); !ok {
		return nil, fmt.Errorf("amount format is invalid: %s", s)
	}
	return res, nil
}

// BigIntToInt64 将 *big.Int 转换为 int64，若超出范围返回错误
func BigIntToInt64(b *big.Int) (int64, error) {
	if b == nil {
		return 0, errors.New("big.Int 不能为 nil")
	}
	// 检查是否在 int64 范围内
	minInt64 := big.NewInt(-(1 << 63))
	maxInt64 := big.NewInt((1 << 63) - 1)
	if b.Cmp(minInt64) < 0 || b.Cmp(maxInt64) > 0 {
		return 0, errors.New("值超出 int64 范围")
	}
	return b.Int64(), nil
}

// Int64ToBigInt 将 int64 转换为 *big.Int
func Int64ToBigInt(n int64) *big.Int {
	return big.NewInt(n)
}
