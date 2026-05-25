package screenshot

import (
	"fmt"
	"os"
)

func ReadBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read screenshot: %w", err)
	}
	return "data:image/png;base64," + encodeBase64(data), nil
}

func Cleanup(path string) {
	os.Remove(path)
}

func encodeBase64(data []byte) string {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	result := make([]byte, 0, (len(data)+2)/3*4)

	for i := 0; i < len(data); i += 3 {
		var buf [3]byte
		n := copy(buf[:], data[i:])

		result = append(result, table[buf[0]>>2])
		result = append(result, table[(buf[0]&0x03)<<4|buf[1]>>4])

		if n > 1 {
			result = append(result, table[(buf[1]&0x0F)<<2|buf[2]>>6])
		} else {
			result = append(result, '=')
		}

		if n > 2 {
			result = append(result, table[buf[2]&0x3F])
		} else {
			result = append(result, '=')
		}
	}

	return string(result)
}
