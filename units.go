/*

 */
package main

import (
	"fmt"
	"github.com/johampel/units/internal"
	"os"
	"strings"
)

var appName = "units"

func showError(message string) {
	_, _ = fmt.Fprintf(os.Stderr, "error %s: %s\n", appName, message)
}

func showErrorAndExit(message string) {
	showError(message)
	os.Exit(1)
}

func usageError() {
	showError("invalid command line")
	os.Exit(1)
}

func unitFile() string {
	dir, _ := os.UserHomeDir()
	return fmt.Sprintf("%s/.units", dir)
}

func showUsage() {
	fmt.Printf("%s\n", appName)
	println("Utility to evaluate physical units.")
	println("When starting the first time, it knows the seven base SI units; you may add further units")
	println("as described below.")
	fmt.Printf("%s has four different sub commands, which are described in the following sections:\n", appName)
	fmt.Printf("%s list\n", appName)
	println("    Prints all known unit definitions; for the seven base SI units it simply prints out the")
	println("    unit symbol, for units created via the 'add' command the symbol plus its definition is")
	println("    shown.")
	fmt.Printf("%s add <unit> <expression>\n", appName)
	println("    Defines a new unit named <unit> using the given <expression>. For example, the following")
	println("    would define an unit named 'km' representing one kilometer:")
	fmt.Printf("      %s add km '1000*m'\n", appName)
	println("    Please refer to the section below for the syntax of <expression>.")
	fmt.Printf("%s remove <unit>\n", appName)
	println("    Removes a unit definition previously added via the 'add' command. A definition cannot be")
	println("    removed, if it is a base unit or it is used for other definitions.")
	fmt.Printf("%s <expression>\n", appName)
	println("    Validates and evaluates the given <expression>. Evaluation basically means that it")
	println("    substitutes all derived types with the according SI base units. For example, if 'km'")
	println("    and 'h' are defined as kilometer resp. hour, the expression '36*km*h^-1` evaluates")
	println("    to '10*m*s^-1'. Please refer to the section below for the syntax of <expression>.")
	println("The general syntax of <expressions> is as follows:")
	println("    [<cofficient>*]<term1>*...*<termN>")
	println("whereas <coefficient> is a floating point number")
	println("and <term> has the format <unit>[^<exponent>]")
}

func list(args []string) {
	if len(args) != 2 {
		usageError()
	}

	if err := internal.LoadUnits(unitFile()); err != nil {
		showError(err.Error())
	}

	names := internal.GetUnitNames()
	for _, name := range names {
		unit, err := internal.GetUnit(name)
		if err != nil {
			showErrorAndExit(err.Error())
		}
		if unit.IsBaseUnit() {
			fmt.Printf("%s\n", name)
		} else {
			fmt.Printf("%s = %s\n", name, unit.GetFormula())
		}
	}
}

func addUnit(args []string) {
	if len(args) != 4 {
		usageError()
	}

	if err := internal.LoadUnits(unitFile()); err != nil {
		showErrorAndExit(err.Error())
	}

	name := args[2]
	formula := args[3]

	if _, err := internal.GetUnit(name); err == nil {
		showErrorAndExit(fmt.Sprintf("Unit '%s' already defined", name))
	}

	expr, err := internal.ParseExpression(formula)
	if err != nil {
		showErrorAndExit(err.Error())
	}
	if err := expr.Validate(); err != nil {
		showErrorAndExit(err.Error())
	}

	if _, err := internal.AddUnit(name, formula); err != nil {
		showErrorAndExit(err.Error())
	}

	if err := internal.StoreUnits(unitFile()); err != nil {
		showErrorAndExit(err.Error())
	}
}

func removeUnit(args []string) {
	if len(args) != 3 {
		usageError()
	}

	if err := internal.LoadUnits(unitFile()); err != nil {
		showErrorAndExit(err.Error())
	}

	name := args[2]
	unit, err := internal.GetUnit(name)
	if err != nil {
		showErrorAndExit(err.Error())
	}
	if unit.IsBaseUnit() {
		showErrorAndExit("Cannot delete a base unit")
	}

	for _, unitName := range internal.GetUnitNames() {
		if unitName == name {
			continue
		}

		unit, err := internal.GetUnit(unitName)
		if err != nil {
			showErrorAndExit(err.Error())
		}
		expr, err := internal.ParseExpression(unit.GetFormula())
		if err != nil {
			showErrorAndExit(err.Error())
		}
		if expr.RefersToUnit(name) {
			showErrorAndExit(fmt.Sprintf("Unit '%s' still in use (at least by '%s')", name, unitName))
		}
	}

	if err := internal.RemoveUnit(name); err != nil {
		showErrorAndExit(err.Error())
	}

	if err := internal.StoreUnits(unitFile()); err != nil {
		showErrorAndExit(err.Error())
	}
}

func eval(args []string) {
	if len(args) != 2 {
		usageError()
	}

	if err := internal.LoadUnits(unitFile()); err != nil {
		showErrorAndExit(err.Error())
	}

	expr, err := internal.ParseExpression(args[1])
	if err != nil {
		showErrorAndExit(err.Error())
	}

	err = expr.Validate()
	if err != nil {
		showErrorAndExit(err.Error())
	}

	expr, err = expr.ReplaceDerivedUnits()
	if err != nil {
		showErrorAndExit(err.Error())
	}

	expr = expr.Normalize()
	println(expr.String())
}

func main() {
	appName = os.Args[0]
	if pos := strings.LastIndex(appName, "/"); pos != -1 {
		appName = appName[pos+1:]
	}

	if len(os.Args) < 2 {
		showUsage()
		return
	}

	switch os.Args[1] {
	case "list":
		list(os.Args)
	case "add":
		addUnit(os.Args)
	case "remove":
		removeUnit(os.Args)
	default:
		eval(os.Args)
	}
}
