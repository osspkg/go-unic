package node

type Block struct {
	key    *Key
	child  []*Block
	parent *Block
}

func NewBlock() *Block {
	return &Block{
		key:    NewKey(),
		child:  nil,
		parent: nil,
	}
}

func (v *Block) IsRoot() bool {
	return v.parent == nil
}

func (v *Block) HasChild() bool {
	return len(v.child) > 0
}

func (v *Block) Root() (b *Block) {
	b = v
	for {
		if b.parent == nil {
			break
		}
		b = v.parent
	}
	return
}

func (v *Block) PreviousBlock() (b *Block) {
	b = v
	if b.parent != nil {
		b = v.parent
	}
	return
}

func (v *Block) NextBlock() *Block {
	b := NewBlock()
	b.parent = v
	v.child = append(v.child, b)
	return b
}

func (v *Block) Key() *Key {
	return v.key
}

func (v *Block) Child() []*Block {
	return v.child
}
