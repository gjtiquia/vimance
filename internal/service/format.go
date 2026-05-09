package service

import "fmt"

func FormatCents(cents int64) string {
	if cents < 0 {
		abs := -cents
		dollars := abs / 100
		c := abs % 100
		return fmt.Sprintf("-%d.%02d", dollars, c)
	}

	dollars := cents / 100
	c := cents % 100
	return fmt.Sprintf("%d.%02d", dollars, c)
}
