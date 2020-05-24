package im

import (
	"encoding/binary"
	"log"
	"os"
	"testing"
)

func TestA(t *testing.T) {
	infoBytes := make([]byte, 16)
	binary.LittleEndian.PutUint64(infoBytes, uint64(100))
	log.Println(infoBytes)
	partBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(partBytes, uint64(30))
	copy(infoBytes[8:], partBytes)
	log.Println(infoBytes)
}

func TestB(t *testing.T) {
	fileInfo, err := os.Stat("/home/m/an.sh")
	if err != nil {
		log.Fatalln("文件路径错误")
	}
	log.Println(fileInfo.Mode())
}
