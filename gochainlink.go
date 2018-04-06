package main

import (
	"net/http"
	"strconv"
	"fmt"
	"encoding/binary"
	"bytes"
	"crypto/sha256"
	"math/big"
	"html/template"
	"log"
)

const difficulty = 2

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

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
var Blockchain []Block

func blkChString() (s string) {
	for _, b := range Blockchain {
		s += b.String()
	}
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

func addBlock(s string) {
	var newBlock Block
	if len(Blockchain) == 0 {
		newBlock.Index = 0
		newBlock.PrevHash = make([]byte, 32)
		newBlock.Difficulty = difficulty
	} else {
		newBlock.Index = Blockchain[len(Blockchain)-1].Index + 1
		newBlock.PrevHash = Blockchain[len(Blockchain)-1].BlockHash
		newBlock.Difficulty = Blockchain[len(Blockchain)-1].Difficulty + difficulty
	}
	newBlock.WkLog.TimeSpent, _ = strconv.Atoi(s)
	newBlock.GenerateHash()
}

func addBlockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	s := r.FormValue("wklog")
	addBlock(s)
	d := struct{
		Wklog string
	}{
		Wklog: s,
	}
	//fmt.Print("Template parsing.\n")
	err := tpl.ExecuteTemplate(w, "newblock.gohtml", d)
	if err != nil {
		log.Panic("Template Error: ", err)
	}
}

func blockchainHandler(w http.ResponseWriter, r *http.Request) {
	for _, b := range Blockchain {
		fmt.Fprint(w, b.String())
	}
}

type BlkData struct {
	Index      string
	WkLog      string
	BlockHash  string
	PrevHash   string
	Difficulty string
	Nonce      string
}

func index(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "index.gohtml", nil)
	if err != nil {
		log.Panic("Template Error: ", err)
	}
}

func chaindata(w http.ResponseWriter, r *http.Request) {
	var d []BlkData
	for _, b := range Blockchain {
		d = append(d, BlkData{strconv.Itoa(b.Index),
			b.WkLog.String(),
			fmt.Sprintf("% x", b.BlockHash),
			fmt.Sprintf("% x", b.PrevHash),
			strconv.Itoa(b.Difficulty),
			fmt.Sprintf("% x", b.Nonce),
		})
	}
	err := tpl.ExecuteTemplate(w, "chaindata.gohtml", d)
	if err != nil {
		log.Panic("Template Error: ", err)
	}
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/addblock", addBlockHandler)
	http.HandleFunc("/chaindata", chaindata)
	http.ListenAndServe(":8080", nil)
}