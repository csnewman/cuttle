package cuttle

type BatchEntry struct {
	Stmt         string
	Args         []any
	ExecHandler  AsyncHandler[Exec]
	QueryHandler AsyncHandler[Rows]
}

type BatchRW struct {
	Entries []*BatchEntry
}

func NewBatchRW() *BatchRW {
	return &BatchRW{}
}

func (b *BatchRW) Exec(handler AsyncHandler[Exec], stmt string, args ...any) {
	b.Entries = append(b.Entries, &BatchEntry{
		Stmt:        stmt,
		Args:        args,
		ExecHandler: handler,
	})
}

func (b *BatchRW) Query(handler AsyncHandler[Rows], stmt string, args ...any) {
	b.Entries = append(b.Entries, &BatchEntry{
		Stmt:         stmt,
		Args:         args,
		QueryHandler: handler,
	})
}

func (b *BatchRW) QueryRow(handler AsyncHandler[Row], stmt string, args ...any) { //nolint:revive
	panic("implement me")
}

type BatchR struct {
	Entries []*BatchEntry
}

func NewBatchR() *BatchR {
	return &BatchR{}
}

func (b *BatchR) Query(handler AsyncHandler[Rows], stmt string, args ...any) {
	b.Entries = append(b.Entries, &BatchEntry{
		Stmt:         stmt,
		Args:         args,
		QueryHandler: handler,
	})
}

func (b *BatchR) QueryRow(handler AsyncHandler[Row], stmt string, args ...any) { //nolint:revive
	panic("implement me")
}

var (
	_ AsyncRTx = (*BatchR)(nil)
	_ AsyncRTx = (*BatchRW)(nil)
	_ AsyncWTx = (*BatchRW)(nil)
)
