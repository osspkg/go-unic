package node

func Search(b *Block, keys ...string) *Block {
	b = b.Root()
	for _, key := range keys {
		var has bool
		for _, block := range b.Child() {
			if key == block.Key().key {
				b = block
				has = true
				break
			}
		}
		if !has {
			return nil
		}
	}
	return b
}
