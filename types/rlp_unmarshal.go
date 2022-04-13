package types

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/umbracle/fastrlp"
)

type RLPUnmarshaler interface {
	UnmarshalRLP(input []byte) error
}

type unmarshalRLPFunc func(p *fastrlp.Parser, v *fastrlp.Value) error

func UnmarshalRlp(obj unmarshalRLPFunc, input []byte) error {
	pr := fastrlp.DefaultParserPool.Get()

	v, err := pr.Parse(input)
	if err != nil {
		fastrlp.DefaultParserPool.Put(pr)

		return err
	}

	if err := obj(pr, v); err != nil {
		fastrlp.DefaultParserPool.Put(pr)

		return err
	}

	fastrlp.DefaultParserPool.Put(pr)

	return nil
}

func (t *TransactionType) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	bytes, err := v.Bytes()
	if err != nil {
		return err
	}

	if l := len(bytes); l != 1 {
		return fmt.Errorf("expected 1 byte transaction type, but size is %d", l)
	}

	if *t, err = ToTransactionType(bytes[0]); err != nil {
		return err
	}

	return nil
}

func (b *Block) UnmarshalRLP(input []byte) error {
	return UnmarshalRlp(b.UnmarshalRLPFrom, input)
}

func (b *Block) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if num := len(elems); num != 3 {
		return fmt.Errorf("not enough elements to decode block, expected 3 but found %d", num)
	}

	// header
	b.Header = &Header{}
	if err := b.Header.UnmarshalRLPFrom(p, elems[0]); err != nil {
		return err
	}

	// transactions
	var txns Transactions
	if err := txns.UnmarshalRLPFrom(p, elems[1]); err != nil {
		return err
	}

	b.Transactions = txns

	// uncles
	uncles, err := elems[2].GetElems()
	if err != nil {
		return err
	}

	for _, uncle := range uncles {
		bUncle := &Header{}
		if err := bUncle.UnmarshalRLPFrom(p, uncle); err != nil {
			return err
		}

		b.Uncles = append(b.Uncles, bUncle)
	}

	return nil
}

func (h *Header) UnmarshalRLP(input []byte) error {
	return UnmarshalRlp(h.UnmarshalRLPFrom, input)
}

func (h *Header) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if num := len(elems); num != 15 {
		return fmt.Errorf("not enough elements to decode header, expected 15 but found %d", num)
	}

	// parentHash
	if err = elems[0].GetHash(h.ParentHash[:]); err != nil {
		return err
	}
	// sha3uncles
	if err = elems[1].GetHash(h.Sha3Uncles[:]); err != nil {
		return err
	}
	// miner
	if err = elems[2].GetAddr(h.Miner[:]); err != nil {
		return err
	}
	// stateroot
	if err = elems[3].GetHash(h.StateRoot[:]); err != nil {
		return err
	}
	// txroot
	if err = elems[4].GetHash(h.TxRoot[:]); err != nil {
		return err
	}
	// receiptroot
	if err = elems[5].GetHash(h.ReceiptsRoot[:]); err != nil {
		return err
	}
	// logsBloom
	if _, err = elems[6].GetBytes(h.LogsBloom[:0], 256); err != nil {
		return err
	}
	// difficulty
	if h.Difficulty, err = elems[7].GetUint64(); err != nil {
		return err
	}
	// number
	if h.Number, err = elems[8].GetUint64(); err != nil {
		return err
	}
	// gasLimit
	if h.GasLimit, err = elems[9].GetUint64(); err != nil {
		return err
	}
	// gasused
	if h.GasUsed, err = elems[10].GetUint64(); err != nil {
		return err
	}
	// timestamp
	if h.Timestamp, err = elems[11].GetUint64(); err != nil {
		return err
	}
	// extraData
	if h.ExtraData, err = elems[12].GetBytes(h.ExtraData[:0]); err != nil {
		return err
	}
	// mixHash
	if err = elems[13].GetHash(h.MixHash[:0]); err != nil {
		return err
	}
	// nonce
	nonce, err := elems[14].GetUint64()
	if err != nil {
		return err
	}

	h.SetNonce(nonce)

	// compute the hash after the decoding
	h.ComputeHash()

	return err
}

func (r *Receipts) UnmarshalRLP(input []byte) error {
	return UnmarshalRlp(r.UnmarshalRLPFrom, input)
}

func (r *Receipts) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	for i := 0; i < len(elems); i++ {
		txType := TxTypeLegacy
		if elems[i].Type() == fastrlp.TypeBytes {
			// Parse Transaction Type if Bytes come first
			if err := txType.UnmarshalRLPFrom(p, elems[i]); err != nil {
				return err
			}

			i++
		}

		rr := &Receipt{
			TransactionType: txType,
		}
		if err := rr.UnmarshalRLPFrom(p, elems[i]); err != nil {
			return err
		}

		*r = append(*r, rr)
	}

	return nil
}

func (r *Receipt) UnmarshalRLP(input []byte) error {
	txType := TxTypeLegacy
	offset := 0

	if len(input) > 0 && input[0] <= RLPSingleByteUpperLimit {
		var err error
		if txType, err = ToTransactionType(input[0]); err != nil {
			return err
		}

		offset = 1
	}

	r.TransactionType = txType
	if err := UnmarshalRlp(r.UnmarshalRLPFrom, input[offset:]); err != nil {
		return err
	}

	return nil
}

// UnmarshalRLP unmarshals a Receipt in RLP format
func (r *Receipt) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if len(elems) != 4 {
		return errors.New("expected 4 elements")
	}

	// root or status
	buf, err := elems[0].Bytes()
	if err != nil {
		return err
	}

	switch size := len(buf); size {
	case 32:
		// root
		copy(r.Root[:], buf[:])
	case 1:
		// status
		r.SetStatus(ReceiptStatus(buf[0]))
	default:
		r.SetStatus(0)
	}

	// cumulativeGasUsed
	if r.CumulativeGasUsed, err = elems[1].GetUint64(); err != nil {
		return err
	}
	// logsBloom
	if _, err = elems[2].GetBytes(r.LogsBloom[:0], 256); err != nil {
		return err
	}

	// logs
	logsElems, err := v.Get(3).GetElems()
	if err != nil {
		return err
	}

	for _, elem := range logsElems {
		log := &Log{}
		if err := log.UnmarshalRLPFrom(p, elem); err != nil {
			return err
		}

		r.Logs = append(r.Logs, log)
	}

	return nil
}

func (l *Log) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if len(elems) != 3 {
		return fmt.Errorf("bad elems")
	}

	// address
	if err := elems[0].GetAddr(l.Address[:]); err != nil {
		return err
	}
	// topics
	topicElems, err := elems[1].GetElems()
	if err != nil {
		return err
	}

	l.Topics = make([]Hash, len(topicElems))

	for indx, topic := range topicElems {
		if err := topic.GetHash(l.Topics[indx][:]); err != nil {
			return err
		}
	}

	// data
	if l.Data, err = elems[2].GetBytes(l.Data[:0]); err != nil {
		return err
	}

	return nil
}

func (tt *Transactions) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	txns, err := v.GetElems()
	if err != nil {
		return err
	}

	for i := 0; i < len(txns); i++ {
		txType := TxTypeLegacy
		if txns[i].Type() == fastrlp.TypeBytes {
			if err := txType.UnmarshalRLPFrom(p, txns[i]); err != nil {
				return err
			}

			i++
		}

		txn := &Transaction{}

		if txn.Payload, err = newTxPayload(txType); err != nil {
			return err
		}

		if err := txn.Payload.UnmarshalRLPFrom(p, txns[i]); err != nil {
			return err
		}

		txn.ComputeHash()

		*tt = append(*tt, txn)
	}

	return nil
}

func (t *Transaction) UnmarshalRLP(input []byte) error {
	txType := TxTypeLegacy
	offset := 0

	var err error
	if len(input) > 0 && input[0] <= RLPSingleByteUpperLimit {
		if txType, err = ToTransactionType(input[0]); err != nil {
			return err
		}

		offset = 1
	}

	if t.Payload, err = newTxPayload(txType); err != nil {
		return err
	}

	if err := UnmarshalRlp(t.Payload.UnmarshalRLPFrom, input[offset:]); err != nil {
		return err
	}

	return nil
}

// UnmarshalRLP unmarshals a Transaction in RLP format
func (t *LegacyTransaction) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if num := len(elems); num != 9 {
		return fmt.Errorf("not enough elements to decode transaction, expected 9 but found %d", num)
	}

	// nonce
	if t.Nonce, err = elems[0].GetUint64(); err != nil {
		return err
	}

	// gasPrice
	t.GasPrice = new(big.Int)
	if err := elems[1].GetBigInt(t.GasPrice); err != nil {
		return err
	}

	// gas
	if t.Gas, err = elems[2].GetUint64(); err != nil {
		return err
	}

	// to
	if vv, _ := v.Get(3).Bytes(); len(vv) == 20 {
		// address
		addr := BytesToAddress(vv)
		t.To = &addr
	} else {
		// reset To
		t.To = nil
	}

	// value
	t.Value = new(big.Int)
	if err := elems[4].GetBigInt(t.Value); err != nil {
		return err
	}

	// input
	if t.Input, err = elems[5].GetBytes(t.Input[:0]); err != nil {
		return err
	}

	// V
	t.V = new(big.Int)
	if err = elems[6].GetBigInt(t.V); err != nil {
		return err
	}

	// R
	t.R = new(big.Int)
	if err = elems[7].GetBigInt(t.R); err != nil {
		return err
	}

	// S
	t.S = new(big.Int)
	if err = elems[8].GetBigInt(t.S); err != nil {
		return err
	}

	return nil
}

func (t *StateTransaction) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if num := len(elems); num != 7 {
		return fmt.Errorf("not enough elements to decode transaction, expected 7 but found %d", num)
	}

	// nonce
	if t.Nonce, err = elems[0].GetUint64(); err != nil {
		return err
	}

	// to
	if vv, _ := v.Get(1).Bytes(); len(vv) == 20 {
		// address
		addr := BytesToAddress(vv)
		t.To = &addr
	} else {
		// reset To
		t.To = nil
	}

	// input
	if t.Input, err = elems[2].GetBytes(t.Input[:0]); err != nil {
		return err
	}

	// Signatures
	sigElems, err := elems[3].GetElems()
	if err != nil {
		return err
	}

	t.Signatures = make([][]byte, len(sigElems))
	for i, e := range sigElems {
		if t.Signatures[i], err = e.GetBytes(t.Signatures[i][:0]); err != nil {
			return err
		}
	}

	// V
	t.V = new(big.Int)
	if err = elems[4].GetBigInt(t.V); err != nil {
		return err
	}

	// R
	t.R = new(big.Int)
	if err = elems[5].GetBigInt(t.R); err != nil {
		return err
	}

	// S
	t.S = new(big.Int)
	if err = elems[6].GetBigInt(t.S); err != nil {
		return err
	}

	return nil
}
