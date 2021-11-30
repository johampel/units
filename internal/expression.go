package internal

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Expression struct {
	coefficient float64
	terms       []term
}

func (t *Expression) Validate() error {
	for _, term := range t.terms {
		if _, err := GetUnit(term.unit); err != nil {
			return err
		}
	}
	return nil
}

func (t *Expression) RefersToUnit(unit string) bool {
	for _, term := range t.terms {
		if term.unit == unit {
			return true
		}
	}
	return false
}

func (t *Expression) ReplaceDerivedUnits() (*Expression, error) {
	terms := make([]term, 0)
	coefficient := t.coefficient

	for _, item := range t.terms {
		unit, err := GetUnit(item.unit)
		if err != nil {
			return nil, err
		}
		if unit.IsBaseUnit() {
			terms = append(terms, item)
			continue
		}

		expr, err := ParseExpression(unit.formula)
		if err != nil {
			return nil, err
		}
		expr, err = expr.ReplaceDerivedUnits()
		if err != nil {
			return nil, err
		}
		for i := range expr.terms {
			expr.terms[i].exponent *= item.exponent
		}
		terms = append(terms, expr.terms...)
		if expr.coefficient != 1.0 {
			coefficient *= math.Pow(expr.coefficient, float64(item.exponent))
		}
	}

	return &Expression{coefficient, terms}, nil
}

func (t *Expression) Normalize() *Expression {
	termsByUnit := make(map[string]term)
	for _, item := range t.terms {
		unit := item.unit
		unitTerm, found := termsByUnit[unit]
		if found {
			unitTerm.exponent += item.exponent
			termsByUnit[unit] = unitTerm
		} else {
			termsByUnit[unit] = term{unit, item.exponent}
		}
	}

	terms := make([]term, 0)
	for _, item := range termsByUnit {
		if item.exponent != 0 {
			terms = append(terms, item)
		}
	}
	return &Expression{t.coefficient, terms}
}

func (t *Expression) String() string {
	if len(t.terms) == 0 {
		return "1"
	}
	builder := strings.Builder{}
	if t.coefficient != 1 {
		builder.WriteString(fmt.Sprintf("%f", t.coefficient))
	}
	for _, term := range t.terms {
		if builder.Len() > 0 {
			builder.WriteString("*")
		}
		builder.WriteString(term.String())
	}
	return builder.String()
}

func ParseExpression(str string) (*Expression, error) {
	var err error
	termStrs := strings.SplitN(str, "*", -1)
	terms := make([]term, 0)
	coeff := 1.0
	for i, termStr := range termStrs {
		if i == 0 {
			coeff, err = strconv.ParseFloat(strings.TrimSpace(termStr), 64)
			if err == nil {
				continue
			} else {
				coeff = 1.0
			}
		}
		term, err := parseTerm(termStr)
		if err != nil {
			return nil, err
		}

		terms = append(terms, term)
	}

	expr := Expression{coeff, terms}
	return &expr, nil
}

type term struct {
	unit     string
	exponent int
}

func (t term) String() string {
	if t.exponent == 0 {
		return "1"
	}
	if t.exponent == 1 {
		return t.unit
	}
	return fmt.Sprintf("%s^%d", t.unit, t.exponent)
}

func parseTerm(str string) (term, error) {
	tokens := strings.SplitN(str, "^", -1)
	if tokens == nil || len(tokens) > 2 {
		return term{}, errors.New(fmt.Sprintf("Invalid term '%s'", str))
	}
	unit := strings.TrimSpace(tokens[0])
	if len(unit) == 0 {
		return term{}, errors.New(fmt.Sprintf("Invalid term '%s'", str))
	}
	exponent := 1
	if len(tokens) == 2 {
		exp, err := strconv.Atoi(strings.TrimSpace(tokens[1]))
		if err != nil {
			return term{}, errors.New(fmt.Sprintf("Invalid term '%s'", str))
		}
		exponent = exp
	}
	return term{unit, exponent}, nil
}
