package app

type updateManager struct {
	pending []map[Composer]int
}

func (m *updateManager) Add(c Composer, v int) {
	depth := int(c.depth())
	if len(m.pending) <= depth {
		size := max(depth+1, 100)
		pending := make([]map[Composer]int, size)
		copy(pending, m.pending)
		m.pending = pending
	}

	updates := m.pending[depth]
	if updates == nil {
		updates = make(map[Composer]int)
		m.pending[depth] = updates
	}
	updates[c] += v
}

func (m *updateManager) Done(v Composer) {
	depth := v.depth()
	if len(m.pending) <= int(depth) {
		return
	}
	updates := m.pending[depth]
	delete(updates, v)
}

func (m *updateManager) UpdateForEach(do func(Composer)) {
	for _, updates := range m.pending {
		for compo, counter := range updates {
			if counter > 0 {
				do(compo)
			}
			delete(updates, compo)
		}
	}
}
