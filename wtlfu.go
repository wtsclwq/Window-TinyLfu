package wtlfu

type WTinyLfu[K comparable] struct {
}

func (l *WTinyLfu[K]) Put(key K) error {
	return nil
}

func (l *WTinyLfu[K]) Access() error {
	return nil
}

func (l *WTinyLfu[K]) Remove(key K) bool {
	return false
}
