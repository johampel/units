package internal

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type Unit struct {
	name    string
	formula string
}

func (t *Unit) GetName() string {
	return t.name
}

func (t *Unit) GetFormula() string {
	return t.formula
}

func (t *Unit) IsBaseUnit() bool {
	return t.formula == t.name
}

func (t *Unit) String() string {
	return t.name
}

var units = make(map[string]*Unit)

func GetUnit(name string) (*Unit, error) {
	unit, found := units[name]
	if !found {
		return nil, errors.New(fmt.Sprintf("Unit '%s' not found", name))
	}
	return unit, nil
}

func AddUnit(name string, formula string) (*Unit, error) {
	if _, found := units[name]; found {
		return nil, errors.New(fmt.Sprintf("Unit '%s' already defined", name))
	}
	unit := &Unit{name, formula}
	units[name] = unit
	return unit, nil
}

func RemoveUnit(name string) error {
	if _, found := units[name]; !found {
		return errors.New(fmt.Sprintf("Unit '%s' not found", name))
	}
	delete(units, name)
	return nil
}

func LoadUnits(unitFile string) error {
	file, err := os.Open(unitFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if scanner.Err() != nil {
			return scanner.Err()
		}
		tokens := strings.SplitN(text, "=", 2)
		if len(tokens) != 2 {
			continue
		}
		if _, err := AddUnit(strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])); err != nil {
			return err
		}
	}
	return nil
}

func StoreUnits(unitFile string) error {
	file, err := os.Create(unitFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, unit := range units {
		if unit.IsBaseUnit() {
			continue
		}
		if _, err := fmt.Fprintf(file, "%s=%s\n", unit.name, unit.formula); err != nil {
			return err
		}
	}
	return nil
}

func GetUnitNames() []string {
	names := make([]string, 0)
	for name := range units {
		names = append(names, name)
	}
	return names
}

func init() {
	for _, unit := range []string{"s", "m", "kg", "A", "K", "mol", "cd"} {
		if _, err := AddUnit(unit, unit); err != nil {
			log.Fatal(err)
		}
	}
}
