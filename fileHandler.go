package main

import (
	// "bufio"
	"fmt"
	"math"
	"os"
)

func byteToInt64(buf []byte) (ret int64) {
	bufSize := len(buf)
	if bufSize == 2 {
		ret = int64(buf[0])<<8 + int64(buf[1])
	} else if bufSize == 4 {
		ret = int64(buf[0])<<16 + int64(buf[1])<<32 + int64(buf[2]) + int64(buf[0])<<8
	} else {
		ret = int64(buf[0])
	}
	return ret
}

func ibm2ieee(buf []byte) (ret float64) {
	// gain sign from first bit
	sign := float64(buf[0] >> 7)

	// gain exponent from first byte, last 7 bits
	exp := float64(buf[0]&0x7f) - 64

	// gain maintissa from last 3 bytes
	frac := float64(int32(buf[1])<<16 + int32(buf[2])<<8 + int32(buf[3]))
	frac = frac / math.Pow(2, 24)

	ret = (1 - 2*sign) * (math.Pow(16, exp)) * frac

	return ret
}

func ParseSegyInfo(fl *os.File) (sampleRate int64, traceLength int64, formatCode int64, traceBytes int64, totalTraces int64) {
	buf := make([]byte, 2)
	// 获取采样率
	{
		fl.ReadAt(buf, 3216)
		sampleRate = byteToInt64(buf)
		fmt.Printf("采样率：%d\n", sampleRate)
	}

	// 获取道长度
	{
		fl.ReadAt(buf, 3220)
		traceLength = byteToInt64(buf)
		fmt.Printf("道长度：%d\n", traceLength)
	}

	// 获取格式码
	{
		fl.ReadAt(buf, 3224)
		formatCode = byteToInt64(buf)
		fmt.Printf("格式码：%d\n", formatCode)
	}

	// 获取每道字节数
	{
		traceBytes = 240 + traceLength*4
		fmt.Printf("每道字节数：%d\n", traceBytes)
	}

	// 获取总道数
	{
		fileInfo, err := fl.Stat()
		if err != nil {

		}
		fileSize := fileInfo.Size()
		totalTraces = (fileSize - 3200 - 400) / traceBytes
		fmt.Printf("总道数：%d\n", totalTraces)
	}

	return sampleRate, traceLength, formatCode, traceBytes, totalTraces
}

func GetHeaderInfo(idx int64, traceBytes int64, fl *os.File) (paoIndex int64, traceIndex int64) {
	buf := make([]byte, 4)
	fl.ReadAt(buf, 3600+idx*traceBytes+9)
	paoIndex = byteToInt64(buf)
	fl.ReadAt(buf, 3600+idx*traceBytes+13)
	traceIndex = byteToInt64(buf)
	return paoIndex, traceIndex
}

func OutputTraces(start int64, count int64, outputCount int, traceBytes int64, fl *os.File) {
	buf := make([]byte, 4)
	for idx := int64(0); idx < count; idx++ {
		paoIndex, traceIndex := GetHeaderInfo(idx, traceBytes, fl)
		fmt.Printf("炮号：%d\t道号：%d\n", paoIndex, traceIndex)
		fmt.Println("------------------------------")

		offset := int64(3600) + idx*traceBytes + 240
		ret, err := fl.Seek(offset, 0)
		if ret != offset || err != nil {
			fmt.Println(err)
			break
		}
		for idn := 0; idn < outputCount; idn++ {
			fl.Read(buf)
			fmt.Println(ibm2ieee(buf))
		}
		fmt.Println("")
	}
}

func AverageEnergyPerTrace(idx int64, traceLength int64, traceBytes int64, fl *os.File) (paoIndex int64, traceIndex int64, energy float64) {
	paoIndex, traceIndex = GetHeaderInfo(idx, traceBytes, fl)

	energy = 0
	offset := int64(3600) + idx*traceBytes + 240
	ret, err := fl.Seek(offset, 0)
	if ret != offset || err != nil {
		fmt.Println(err)
		return
	}
	buf := make([]byte, 4)
	for idx := int64(0); idx < traceLength; idx++ {
		fl.Read(buf)
		energy += math.Pow(ibm2ieee(buf), 2) / float64(traceLength)
	}
	fmt.Printf("炮号：%-3d\t道号：%-3d 平均能量: %f\n", paoIndex, traceIndex, energy)

	return paoIndex, traceIndex, energy
}

/*
////////////////////////////////////////////////////////////////////////////////

func sampleReadln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func sampleRead() {
	inFile := "./test/energy.txt"
	fi, err := os.Open(inFile)
	defer fi.Close()
	if err != nil {
		fmt.Println(inFile, err)
		return
	}

	var paoIndex int64
	var traceIndex int64
	var energy float64
	reader := bufio.NewReader(fi)
	for {
		content, err := sampleReadln(reader)
		if err != nil {
			break
		}
		// fmt.Println(content)
		fmt.Sscanf(content, "炮号=%d 道号=%d 平均能量=%f\n", &paoIndex, &traceIndex, &energy)
		fmt.Printf("炮号=%-3d 道号=%-3d 平均能量=%f\n", paoIndex, traceIndex, energy)
	}

}

func sampleWrite() {
	userFile := "example2d.sgy"
	fl, err := os.Open(userFile)
	defer fl.Close()
	if err != nil {
		fmt.Println(userFile, err)
		return
	}

	fmt.Println("==================================================")
	_, traceLength, _, traceBytes, _ := parseSegyInfo(fl)
	fmt.Println("==================================================\n")
	// outputTraces(0, 2000, 0, traceBytes, fl)

	outFile := "./test/energy.txt"
	fo, err1 := os.Create(outFile)
	defer fo.Close()
	if err1 != nil {
		fmt.Println(outFile, err1)
		return
	}

	for idx := int64(0); idx < 100; idx++ {
		paoIndex, traceIndex, energy := averageEnergyPerTrace(idx, traceLength, traceBytes, fl)
		fmt.Fprintf(fo, "炮号=%-3d 道号=%-3d 平均能量=%f\n", paoIndex, traceIndex, energy)
	}

}

func sampleCalculate() {
	inFile := "./test/energy.txt"
	fi, err := os.Open(inFile)
	defer fi.Close()
	if err != nil {
		fmt.Println(inFile, err)
		return
	}

	outFile := "./test/result.txt"
	fo, err := os.Create(outFile)
	defer fo.Close()
	if err != nil {
		fmt.Println(outFile, err)
		return
	}

	var paoIndex int64
	var traceIndex int64
	var energy float64
	var propCoef float64
	reader := bufio.NewReader(fi)
	for {
		content, err := sampleReadln(reader)
		if err != nil {
			break
		}
		// fmt.Println(content)
		fmt.Sscanf(content, "炮号=%d 道号=%d 平均能量=%f\n", &paoIndex, &traceIndex, &energy)
		if energy == 0 {
			propCoef = 0
		} else {
			propCoef = 10000.0 / energy
		}
		fmt.Fprintf(fo, "炮号=%-3d 道号=%-3d 比例系数=%f\n", paoIndex, traceIndex, propCoef)
	}
}

func main() {
	// sampleWrite()
	// sampleRead()
	// sampleCalculate()
}
*/
