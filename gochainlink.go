package main

import (
	"net/http"
	"strconv"
	"fmt"
	"encoding/binary"
	"bytes"
	"crypto/sha256"
	"math/big"
)

const difficulty = 2

type WorkLog struct {
	TimeSpent int // in mins
	//To be added: Project   string
}

func (w *WorkLog) Bytes() ([]byte) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b ,uint64(w.TimeSpent))
	return b
}

func (w *WorkLog) String() (s string) {
	s = fmt.Sprint(w.TimeSpent)
	return s
}

type Block struct {
	Index      int
	WkLog      WorkLog
	BlockHash  []byte
	PrevHash   []byte
	Difficulty int
	Nonce      []byte
}

var Blockchain []Block

func (b *Block) HashValid() (v bool) {
	return false
}

func (b *Block) GenerateHash() {
	i := make([]byte, 8)
	binary.LittleEndian.PutUint64(i, uint64(b.Index))
	buf := bytes.NewBuffer(i)
	_,_ = buf.Write(b.WkLog.Bytes())
	_,_ = buf.Write(b.BlockHash)
	_,_ = buf.Write(b.PrevHash)
	d := make([]byte, 8)
	binary.LittleEndian.PutUint64(d, uint64(b.Difficulty))
	_,_ = buf.Write(d)
	//fmt.Printf("% x\n", buf.Bytes())

	check := bytes.Repeat([]byte{0}, difficulty)
	ind := big.NewInt(0)
	one := big.NewInt(1)
	for {
		hash := sha256.Sum256(append(buf.Bytes(),ind.Bytes()...))
		b.BlockHash = hash[:]
		if bytes.HasPrefix(b.BlockHash, check) {
			fmt.Print("Found Correct Nonce: ", ind, "\n")
			b.Nonce = ind.Bytes()
			Blockchain = append(Blockchain, *b)
			return
		}
		fmt.Print("Tried Incorrect Nonce: ", ind, "\n")
		ind.Add(ind, one)
	}
}

func (b *Block) String() (s string) {
	s = fmt.Sprint("Index: ", b.Index, "\n")
	s += fmt.Sprint("WkLog: ", b.WkLog.String(), "\n")
	s += fmt.Sprintf("BlockHash: % x\n", b.BlockHash)
	s += fmt.Sprintf("PrevHash: % x\n", b.PrevHash)
	s += fmt.Sprint("Difficulty: ", b.Difficulty, "\n")
	s += fmt.Sprintf("Nonce: % x\n\n", b.Nonce)
	return s
}

func addBlockHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "favicon.ico" {return}

	var newBlock Block
	fmt.Print(r.URL.Path[1:], "\n")

	if len(Blockchain) == 0 {
		newBlock.Index = 0
		newBlock.PrevHash = make([]byte, 32)
		newBlock.Difficulty = difficulty
	} else {
		newBlock.Index = Blockchain[len(Blockchain)-1].Index + 1
		newBlock.PrevHash = Blockchain[len(Blockchain)-1].BlockHash
		newBlock.Difficulty = Blockchain[len(Blockchain)-1].Difficulty + difficulty
	}
	newBlock.WkLog.TimeSpent, _ = strconv.Atoi(r.URL.Path[1:])
	newBlock.GenerateHash()
}

func blockchainHandler(w http.ResponseWriter, r *http.Request) {
	for _, b := range Blockchain {
		fmt.Fprint(w, b.String())
	}
}

func main() {
	http.HandleFunc("/", addBlockHandler)
	http.HandleFunc("/blockchain", blockchainHandler)
	http.ListenAndServe(":8080", nil)
}