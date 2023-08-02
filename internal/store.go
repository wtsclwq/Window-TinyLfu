package internal

type RemoveReason uint8

const (
	REMOVED RemoveReason = iota
	EVICTED
	EXPIRED
)
