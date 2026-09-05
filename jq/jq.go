package jq

import (
	"fmt"

	"nikand.dev/go/json2"
)

type (
	// Filter is a general filter interface.
	// Filter parses single value from r buffer starting from st index,
	// processes it, appends result to the w buffer and returns it as the first return parameter.
	//
	// Filters are stateless,
	// but some of them may have temporary buffer they use for internal processing.
	// If a filter needs to carry a state, it returns it from the first call and
	// expects to receive it back on the next call.
	// The first time the filter is called on a new value state must be nil.
	//
	// Filter may not add anything to w buffer if there is an empty result,
	// iterating over empty array is an example.
	//
	// Filter also may have more than one result,
	// iterating over array with more than one element is an example.
	// In that case Next returns only one value at a time and non-nil state.
	// Non-nil returned state means there are more values possible, but not guaranteed.
	// To get them call the Next once again passing returned index and state back to the Filter.
	// Thus the returned index is a part of the state.
	// If returned state is nil, there are no more values to return.
	//
	// Filter may add no value to w and return non-nil state as many times as it needs.
	Filter interface {
		Next(w, r []byte, st int, state State) ([]byte, int, State, error)
	}

	// State is a state of a Filter stored externally as Filters are stateless by design.
	//
	// It's opaque value for the caller and only should be used to pass it back
	// to the same Filter with the same buffer and index returned with the state.
	State any

	// Dot is a trivial filter that parses the values and returns it unchanged.
	Dot struct{}

	// Empty is a filter returning nothing on any input value.
	// It parses the input value though.
	Empty struct{}

	// Literal is a filter returning the same values on any input.
	// It parses the input value though.
	Literal []byte

	// First returns at most one value of the input.
	First struct{}

	Func func(w, r []byte, st int, state State) ([]byte, int, State, error)

	Dumper func(w, r []byte, st, end int)

	// ParseError contains error and position returned by json2.Iterator.
	ParseError struct {
		Err error
		Pos int
	}
)

// ApplyToAll applies the filter f to all the values from the input buffer
// and returns all the results separated by sep sequence ('\n' by default).
//
// NextAll reads only the first value from the input and processes it
// while ApplyToAll does that for each input value until end of buffer reached.
func ApplyToAll(f Filter, w, r, sep []byte) ([]byte, error) {
	var err error

	wend := len(w)

	for i := json2.SkipSpaces(r, 0); i < len(r); i = json2.SkipSpaces(r, i) {
		if wend != len(w) {
			w = append(w, sep...)
		}

		wend = len(w)

		w, i, err = NextAll(f, w, r, i, sep)
		if err != nil {
			return w, err
		}
	}

	return w, nil
}

// NextAll applies filter to the first value in input and return all the result values.
//
// NextAll reads only the first value from the input and processes it
// while ApplyToAll does that for each input value until end of buffer reached.
func NextAll(f Filter, w, r []byte, st int, sep []byte) ([]byte, int, error) {
	var err error
	var state State

	if sep == nil {
		sep = []byte{'\n'}
	}

	wst := len(w)

	for {
		if wst != len(w) {
			w = append(w, sep...)
		}

		wst = len(w)

		w, st, state, err = f.Next(w, r, st, state)
		if err != nil || state == nil {
			return w, st, err
		}
	}
}

func (f Dot) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	var p json2.Iterator

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, state, pe(err, i)
	}

	w = append(w, raw...)

	return w, i, nil, nil
}

func (f Empty) Next(w, r []byte, st int, _ State) (_ []byte, i int, _ State, err error) {
	var p json2.Iterator

	i, err = p.Skip(r, st)

	return w, i, nil, err
}

func (f Literal) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json2.Iterator

	i, err = p.Skip(r, st)
	if err != nil {
		return w, i, state, err
	}

	w = append(w, f...)

	return w, i, nil, nil
}

func (f First) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	var p json2.Iterator

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, state, pe(err, i)
	}

	for i = p.SkipSpaces(r, i); i < len(r); i = p.SkipSpaces(r, i) {
		i, err = p.Skip(r, i)
		if err != nil {
			return w, i, state, pe(err, i)
		}
	}

	w = append(w, raw...)

	return w, i, nil, nil
}

func (f Func) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	return f(w, r, st, state)
}

func (f Dumper) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	st = json2.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	i, err := (&json2.Iterator{}).Skip(r, st)
	if err != nil {
		return w, st, state, err
	}

	f(w, r, st, i)

	return w, i, nil, nil
}

func pe(err error, i int) error {
	//	tlog.Printw("parse error", "i", i, "err", err, "from", loc.Callers(1, 3))
	return ParseError{Err: err, Pos: i}
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse input: %v (at pos %d)", e.Err.Error(), e.Pos)
}

func (e ParseError) Unwrap() error { return e.Err }

func cint(c bool, x, y int) int {
	if c {
		return x
	}

	return y
}

func cbuf(c bool, x, y []byte) []byte {
	if c {
		return x
	}

	return y
}

func cfilter(f0, f1 Filter) Filter {
	if f0 != nil {
		return f0
	}

	return f1
}
