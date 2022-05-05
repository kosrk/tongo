package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"os"
	"strings"
)

type parsedLine struct {
	Text     string
	TypeName string
	Tag      uint32
}

func main() {
	lines, err := readTLFile("liteclient/tmp/lite_api.tl")
	if err != nil {
		panic(err)
	}
	err = writeGoFile(lines)
	if err != nil {
		panic(err)
	}
}

func readTLFile(path string) ([]parsedLine, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var (
		lines []parsedLine
	)
	for scanner.Scan() {
		line := scanner.Text()
		pLine, err := processLine(line)
		if err != nil {
			return nil, err
		}
		if pLine != nil {
			lines = append(lines, *pLine)
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func cleanLine(line string) (string, error) {
	line = strings.Join(strings.Fields(line), " ") // remove extra spaces
	if line == "" {
		return "", nil
	}
	if strings.HasPrefix(line, "---") { // ---text--- delimiter
		return "", nil
	}
	if strings.HasPrefix(line, "//") { // comment line
		return "", nil
	}
	line = strings.SplitAfter(line, ";")[0] // remove data after delimiter
	if !strings.HasSuffix(line, ";") {
		return "", fmt.Errorf("no ; delimiter in row")
	}
	return line[:len(line)-1], nil
}

func processLine(line string) (*parsedLine, error) {
	line, err := cleanLine(line)
	if err != nil {
		return nil, err
	}
	if line == "" {
		return nil, nil
	}
	var pLine parsedLine
	pLine.Text = line
	typeName := strings.SplitAfter(line, " ")[0]
	typeName = typeName[:len(typeName)-1]
	parts := strings.SplitAfter(typeName, "#")
	if len(parts) > 1 {
		pLine.TypeName = convertToConstantName(parts[0][:len(parts[0])-1])
		b, err := hex.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}
		if len(b) != 4 {
			if err != nil {
				return nil, fmt.Errorf("invalid type id")
			}
		}
		i := binary.BigEndian.Uint32(b)
		pLine.Tag = reverseTagBytes(i)
	}
	parts = strings.SplitAfter(typeName, "$")
	if len(parts) > 1 {
		return nil, fmt.Errorf("binary tag not implemented")
		//TODO: implement binary tag
	}
	if pLine.Tag == 0 {
		pLine.TypeName = convertToConstantName(typeName)
		pLine.Tag = reverseTagBytes(crc32.ChecksumIEEE([]byte(pLine.Text)))
	}
	return &pLine, nil
}

func reverseTagBytes(tag uint32) uint32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, tag)
	return binary.LittleEndian.Uint32(b)
}

func writeGoFile(lines []parsedLine) error {
	f, err := os.OpenFile("liteclient/models.go", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	header := `// Code autogenerated. DO NOT EDIT.

package liteclient

const (
`
	_, err = f.Write([]byte(header))
	if err != nil {
		return err
	}
	for _, line := range lines {
		t := fmt.Sprintf("	%v uint32 = 0x%x\n", line.TypeName, line.Tag)
		_, err = f.Write([]byte(t))
		if err != nil {
			return err
		}
	}
	footer := `)`
	_, err = f.Write([]byte(footer))
	if err != nil {
		return err
	}
	return nil
}

func convertToConstantName(text string) string {
	res := ""
	parts := strings.Split(text, ".")
	for _, t := range parts {
		res = res + strings.ToUpper(string(t[0])) + t[1:]
	}
	return res + "Tag"
}
