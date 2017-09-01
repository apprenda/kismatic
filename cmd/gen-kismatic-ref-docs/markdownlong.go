package main

import (
	"fmt"
	"strings"
)

type markdown struct{}

func (markdown) render(docs []doc) {
	fmt.Println("# Plan File Reference")
	fmt.Println("## Index")
	fmt.Println("| Property | Kind | Options |")
	fmt.Println("|----------|------|---------|")
	for _, d := range docs {
		if isStruct(d.propertyType) {
			continue
		}
		fmt.Printf("| %s | %s | %s |\n", d.property, d.propertyType, strings.Join(d.options, ","))
	}
	// for _, d := range docs {
	// 	props := strings.Split(d.property, ".")
	// 	spaces := strings.Repeat("  ", len(props)-1)
	// 	link := strings.Replace(d.property, ".", "", -1)
	// 	prop := props[len(props)-1]
	// 	if d.deprecated {
	// 		prop = prop + " _(deprecated)_"
	// 		link = link + "-deprecated"
	// 	}
	// 	fmt.Printf("%s* [%s](#%s)\n", spaces, prop, link)
	// }

	for _, d := range docs {
		if d.deprecated {
			fmt.Println("###", d.property, "_(deprecated)_")
		} else {
			fmt.Println("###", d.property)
		}
		fmt.Println()
		fmt.Println(d.description)
		fmt.Println()
		fmt.Println("| | |")
		fmt.Println("|----------|-----------------|")
		fmt.Println("| **Kind** | ", d.propertyType, "|")
		req := "No"
		if d.required {
			req = "Yes"
		}
		fmt.Println("| **Required** | ", req, "|")
		def := d.defaultValue
		if def == "" && d.propertyType == "bool" {
			def = "false"
		}
		if def == "" {
			def = " "
		}
		fmt.Printf("| **Default** | `%s` | \n", def)
		if len(d.options) > 0 {
			fmt.Println("| **Options** | ", strings.Join(wrapEach(d.options, "`"), ", "))
		}
		fmt.Println()
	}
}

func wrapEach(s []string, w string) []string {
	res := make([]string, len(s))
	for i := 0; i < len(s); i++ {
		res[i] = w + s[i] + w
	}
	return res
}
