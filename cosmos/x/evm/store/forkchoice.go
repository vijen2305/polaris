package store

import "cosmossdk.io/store"

type Forkchoice struct {
	store store.KVStore
}

func NewForkchoice(store store.KVStore) *Forkchoice {
	return &Forkchoice{
		store: store,
	}
}

func (f *Forkchoice) FinalizedPayloadBlockHash() [32]byte {
	return [32]byte(f.store.Get([]byte("finalizedPayloadBlockHash")))
}

func (f *Forkchoice) SetFinalizedPayloadBlockHash(hash [32]byte) {
	f.store.Set([]byte("finalizedPayloadBlockHash"), hash[:])
}
