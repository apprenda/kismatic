package main

import (
	"fmt"
	"strings"
)

type markdownTable struct{}

func (mt markdownTable) render(docs []doc) {
	fmt.Println("| Property | Type | Required | Default Value | Allowed Value | Description |")
	fmt.Println("|----------|------|----------|---------------|---------------|-------------|")
	for _, d := range docs {
		options := strings.Join(d.options, ",")
		defaultVal := d.defaultValue
		if defaultVal == "" {
			defaultVal = " "
		}
		fmt.Printf("| `%s` | %s | %v | `%s` | %v | %s |\n", d.property, d.propertyType, d.required, defaultVal, options, d.description)
	}
}
