package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

// copyed from internet
func GetMD5SumString(fn string) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", err
	}

	fSum := md5.New()
	_, err = io.Copy(fSum, f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%X", fSum.Sum(nil)), nil
}

func CompareMD5Sum(fileNames []string) (bool, error) {
	sums := []string{}
	for _, fn := range fileNames {
		sum, err := GetMD5SumString(fn)
		if err != nil {
			return false, err
		}
		sums = append(sums, sum)
	}

	lastSum := sums[0]
	for _, sum := range sums[1:] {
		if sum != lastSum {
			return false, fmt.Errorf("Find differ md5 sum")
		}
        lastSum = sum
	}
	return true, nil
}
