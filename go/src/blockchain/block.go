package blockchain

// Transaction is the atomic unit of the legder
type Transaction struct {
	Sender, Recipient string
	Val               float64
}

// Block is a node on the chain and is added once at a time during mining
type Block struct {
	Index    int
	PrevHash int
	Txs      []Transaction
}
