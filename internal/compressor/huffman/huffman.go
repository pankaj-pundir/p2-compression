package huffman

import (
	"archive/zip"
	"bytes"
	"container/heap"
	"fmt"
	"io"
	"os"
)

// HuffmanNode represents a node in the Huffman tree.
type HuffmanNode struct {
	Freq  int
	Char  byte
	Left  *HuffmanNode
	Right *HuffmanNode
}

// PriorityQueue implements heap.Interface for HuffmanNode.
type PriorityQueue []*HuffmanNode

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Freq < pq[j].Freq
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*HuffmanNode)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// buildFrequencyMap builds a frequency map of bytes in the input data.
func buildFrequencyMap(data []byte) map[byte]int {
	freqMap := make(map[byte]int)
	for _, b := range data {
		freqMap[b]++
	}
	return freqMap
}

// buildHuffmanTree builds the Huffman tree from the frequency map.
func buildHuffmanTree(freqMap map[byte]int) *HuffmanNode {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	for char, freq := range freqMap {
		heap.Push(&pq, &HuffmanNode{Char: char, Freq: freq})
	}

	for pq.Len() > 1 {
		left := heap.Pop(&pq).(*HuffmanNode)
		right := heap.Pop(&pq).(*HuffmanNode)
		newNode := &HuffmanNode{Freq: left.Freq + right.Freq, Left: left, Right: right}
		heap.Push(&pq, newNode)
	}

	return heap.Pop(&pq).(*HuffmanNode)
}

// buildCodeMap builds the Huffman code map from the Huffman tree.
func buildCodeMap(root *HuffmanNode, code string, codeMap map[byte]string) {
	if root == nil {
		return
	}
	if root.Left == nil && root.Right == nil {
		codeMap[root.Char] = code
		return
	}
	buildCodeMap(root.Left, code+"0", codeMap)
	buildCodeMap(root.Right, code+"1", codeMap)
}

// compressData compresses the data using the Huffman code map.
func compressData(data []byte, codeMap map[byte]string) []byte {
	var compressed bytes.Buffer
	for _, b := range data {
		compressed.WriteString(codeMap[b])
	}

	// Pad the last byte if necessary
	padding := 8 - (compressed.Len() % 8)
	for i := 0; i < padding; i++ {
		compressed.WriteString("0")
	}

	// Convert the binary string to bytes
	compressedBytes := make([]byte, compressed.Len()/8)
	for i := 0; i < compressed.Len(); i += 8 {
		byteString := compressed.String()[i : i+8]
		byteValue := 0
		for j := 0; j < 8; j++ {
			if byteString[j] == '1' {
				byteValue |= (1 << (7 - j))
			}
		}
		compressedBytes[i/8] = byte(byteValue)
	}

	// Prepend the padding information
	result := make([]byte, 1)
	result[0] = byte(padding)
	result = append(result, compressedBytes...)

	return result
}

// HuffmanCompressor implements the Compressor interface for Huffman Coding.
type HuffmanCompressor struct{}

// Compress compresses the input file using Huffman Coding.
func (c *HuffmanCompressor) Compress(inputFile string, outputFile string, params map[string]interface{}) error {
	// Read the input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Build frequency map
	freqMap := buildFrequencyMap(data)

	// Build Huffman tree
	root := buildHuffmanTree(freqMap)

	// Build code map
	codeMap := make(map[byte]string)
	buildCodeMap(root, "", codeMap)

	// Compress the data
	compressedData := compressData(data, codeMap)

	// Write the compressed data and the code map to a zip file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)

	// Write the code map
	codeMapFile, err := zipWriter.Create("code_map.txt")
	if err != nil {
		return fmt.Errorf("failed to create code map entry in zip: %w", err)
	}
	for char, code := range codeMap {
		_, err := codeMapFile.Write([]byte(fmt.Sprintf("%d:%s\n", char, code)))
		if err != nil {
			return fmt.Errorf("failed to write code map entry: %w", err)
		}
	}

	// Write the compressed data
	compressedFile, err := zipWriter.Create("compressed_data.bin")
	if err != nil {
		return fmt.Errorf("failed to create compressed data entry in zip: %w", err)
	}
	_, err = compressedFile.Write(compressedData)
	if err != nil {
		return fmt.Errorf("failed to write compressed data: %w", err)
	}

	err = zipWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to close zip writer: %w", err)
	}

	return nil
}

// Decompress decompresses the input file using Huffman Coding.
func (c *HuffmanCompressor) Decompress(inputFile string, outputFile string) error {
	// Read the compressed zip file
	r, err := zip.OpenReader(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open compressed file: %w", err)
	}
	defer r.Close()

	var codeMapFile *zip.File
	var compressedDataFile *zip.File

	for _, f := range r.File {
		if f.Name == "code_map.txt" {
			codeMapFile = f
		} else if f.Name == "compressed_data.bin" {
			compressedDataFile = f
		}
	}

	if codeMapFile == nil || compressedDataFile == nil {
		return fmt.Errorf("compressed file is missing code map or compressed data")
	}

	// Read the code map
	codeMapReader, err := codeMapFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open code map entry in zip: %w", err)
	}
	defer codeMapReader.Close()

	codeMapData, err := io.ReadAll(codeMapReader)
	if err != nil {
		return fmt.Errorf("failed to read code map data: %w", err)
	}

	reverseCodeMap := make(map[string]byte)
	lines := bytes.Split(codeMapData, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := bytes.Split(line, []byte(":"))
		if len(parts) != 2 {
			return fmt.Errorf("invalid code map entry: %s", string(line))
		}
		charByte := parts[0][0]
		code := string(parts[1])
		reverseCodeMap[code] = charByte
	}

	// Read the compressed data
	compressedDataReader, err := compressedDataFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open compressed data entry in zip: %w", err)
	}
	defer compressedDataReader.Close()

	compressedData, err := io.ReadAll(compressedDataReader)
	if err != nil {
		return fmt.Errorf("failed to read compressed data: %w", err)
	}

	if len(compressedData) < 1 {
		return fmt.Errorf("compressed data is empty")
	}

	padding := int(compressedData[0])
	compressedBytes := compressedData[1:]

	// Convert bytes to binary string
	var binaryString bytes.Buffer
	for _, b := range compressedBytes {
		binaryString.WriteString(fmt.Sprintf("%08b", b))
	}

	// Remove padding
	if padding > 0 {
		binaryString.Truncate(binaryString.Len() - padding)
	}


	// Decompress the data
	var decompressedData bytes.Buffer
	currentCode := ""
	for _, bit := range binaryString.String() {
		currentCode += string(bit)
		if char, ok := reverseCodeMap[currentCode]; ok {
			decompressedData.WriteByte(char)
			currentCode = ""
		}
	}

	// Write the decompressed data to the output file
	err = os.WriteFile(outputFile, decompressedData.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}