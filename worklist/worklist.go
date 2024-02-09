package worklist

type Entry struct {
	Path string
}

type Worklist struct {
	jobs chan Entry
}

// (w *Worklist) specifies that the Add method is a method of the Worklist type
func (w *Worklist) Add(work Entry) {
	w.jobs <- work
}

func (w *Worklist) Next() Entry {
	j := <-w.jobs
	return j
}

func New(bufSize int) Worklist {
	return Worklist{make(chan Entry, bufSize)}
}

func NewJob(path string) Entry {
	return Entry{path}
}

// Terminate workers by passing empty path to each
func (w *Worklist) Finalize(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		w.Add(Entry{""})
	}
}
