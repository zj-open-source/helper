package datatypes

import (
	"bytes"
	"encoding"
)

type Textable interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	IsZero() bool
}

// inspire by https://www.logicbig.com/tutorials/misc/groovy/range-operator.html
type ValueOrRangeOpt struct {
	ExclusiveFrom bool
	ExclusiveTo   bool
	Exactly       bool
}

func (valueOrRange *ValueOrRangeOpt) UnmarshalText(text []byte, fromValue Textable, toValue Textable) error {
	if len(text) == 0 {
		return nil
	}

	r := ValueOrRangeOpt{}

	spliter := []byte("..")

	r.Exactly = !bytes.Contains(text, spliter)

	fromTo := bytes.Split(text, spliter)

	if len(fromTo) > 0 {
		from := fromTo[0]

		if len(from) > 0 {
			lastChar := from[len(from)-1]
			if lastChar == '!' || lastChar == '<' {
				from = from[:len(from)-1]
				r.ExclusiveFrom = true
			}
		}

		if len(from) > 0 {
			err := fromValue.UnmarshalText(from)
			if err != nil {
				return err
			}
		}
	}

	if len(fromTo) > 1 {
		to := fromTo[1]

		if len(to) > 0 {
			firstChar := to[0]
			if firstChar == '!' || firstChar == '<' {
				to = to[1:]
				r.ExclusiveTo = true
			}
		}

		if len(to) > 0 {
			err := toValue.UnmarshalText(to)
			if err != nil {
				return err
			}
		}
	}

	*valueOrRange = r

	return nil
}

func (valueOrRange ValueOrRangeOpt) MarshalText(fromValue Textable, toValue Textable) (text []byte, err error) {
	buf := bytes.NewBuffer(nil)

	if !fromValue.IsZero() {
		from, err := fromValue.MarshalText()
		if err != nil {
			return nil, err
		}

		buf.Write(from)
		if valueOrRange.ExclusiveFrom {
			buf.WriteByte('<')
		}
	}

	if !valueOrRange.Exactly {
		buf.WriteString("..")

		if !toValue.IsZero() {
			if valueOrRange.ExclusiveTo {
				buf.WriteByte('<')
			}
			to, err := toValue.MarshalText()
			if err != nil {
				return nil, err
			}

			buf.Write(to)
		}
	}

	return buf.Bytes(), nil
}
