package db

import (
	"strconv"
)

func convertFormatUintAppend(data []uint64) string {
	t := make([]byte, 0, len(data)*2)
	for _, a := range data {
		t = strconv.AppendUint(t, a, 10)
		t = append(t, ',')
	}
	return string(t[0 : len(t)-1])
}
