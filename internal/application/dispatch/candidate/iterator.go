package candidate

type Iterator interface {
	Next() (*Candidate, bool)
}
