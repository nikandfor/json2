package json2

type RawMessage []byte

func (x RawMessage) IsZero() bool { return len(x) == 0 }

func (x RawMessage) MarshalJSON() ([]byte, error) {
	if len(x) == 0 {
		return []byte("null"), nil
	}

	return x, nil
}

func (x *RawMessage) UnmarshalJSON(d []byte) error {
	*x = append((*x)[:0], d...)

	return nil
}
