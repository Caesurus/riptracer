package riptracer

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
)

func findNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}
	return results
}

func parseNumbers(input string) ([]int, error) {
	strNums := strings.Fields(input)
	nums := make([]int, 0, len(strNums))
	for _, strNum := range strNums {
		num, err := strconv.Atoi(strNum)
		if err != nil {
			return nil, fmt.Errorf("Error parsing number: %v", err)
		}
		nums = append(nums, num)
	}
	return nums, nil
}

func Dump(buff []byte) {
	n := len(buff)
	rowcount := 0
	stop := (n / 16) * 16
	cnt := 0
	for i := 0; i <= stop; i += 16 {
		cnt++

		if i+16 <= n {
			rowcount = 16
		} else {
			rowcount = min(cnt*16, n) % 16
			if 0 == rowcount {
				break
			}
		}

		// Print offset
		fmt.Printf("0x%04x:  %s", i, Green)

		// Print hex
		for j := 0; j < rowcount; j++ {
			fmt.Printf("%02x ", buff[i+j])
			if j == (rowcount/2)-1 {
				fmt.Printf(" ")
			}
		}
		fmt.Printf("  ")

		fmt.Printf(Reset)
		fmt.Printf("  '%s'\n", viewString(buff[i:(i+rowcount)]))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func viewString(b []byte) string {
	r := []rune(string(b))
	for i := range r {
		if r[i] < 32 || r[i] > 126 {
			r[i] = '.'
		}
	}
	return string(r)
}

func uintptrToBytes(ptr uintptr) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(ptr))
	return bytes
}

func bytesToUint64(b []byte) uint64 {
	var val uint64
	header := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	valPtr := (*uint64)(unsafe.Pointer(header.Data))
	val = *valPtr
	return val
}
