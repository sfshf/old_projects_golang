package util

import "crypto/md5"

func MD5(data, prefix []byte) ([]byte, error) {
	h := md5.New()
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}
	return h.Sum(prefix), nil
}
