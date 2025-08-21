// heavily based in: https://github.com/google/capslock/blob/5bc619fe5e895cccb4cf97b452b8c3fdfb1e142a/cmd/capslock-git-diff/main.go

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"

	cpb "github.com/google/capslock/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	verbose     = flag.Bool("v", false, "enable verbose logging")
	granularity = flag.String("granularity", "intermediate", "the granularity to use for comparisons")
)

func vlog(format string, a ...any) {
	if !*verbose {
		return
	}
	log.Printf(format, a...)
}

func main() {
	b1, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal("error reading capslock.json")
	}
	cil1 := new(cpb.CapabilityInfoList)
	if err = protojson.Unmarshal(b1, cil1); err != nil {
		log.Fatal("error unmarshalling 1")
	}
	vlog("parsed CapabilityInfoList with %d entries", len(cil1.CapabilityInfo))

	b2, err := os.ReadFile(os.Args[2])
	if err != nil {
		log.Fatal("error reading capslock2.json")
	}
	cil2 := new(cpb.CapabilityInfoList)
	if err = protojson.Unmarshal(b2, cil2); err != nil {
		log.Fatal("error unmarshalling 2")
	}
	vlog("parsed CapabilityInfoList with %d entries", len(cil2.CapabilityInfo))

	different := diffCapabilityInfoLists(cil1, cil2)
	if different {
		log.Println("Different")
	}
}

type mapKey struct {
	key        string
	capability cpb.Capability
}
type capabilitiesMap map[mapKey]*cpb.CapabilityInfo

func populateMap(cil *cpb.CapabilityInfoList, granularity string) capabilitiesMap {
	m := make(capabilitiesMap)
	for _, ci := range cil.GetCapabilityInfo() {
		var key string
		switch granularity {
		case "package", "intermediate":
			key = ci.GetPackageDir()
		case "function", "":
			if len(ci.Path) == 0 {
				continue
			}
			key = ci.Path[0].GetName()
		default:
			panic("unknown granularity " + granularity)
		}
		if key == "" {
			continue
		}
		m[mapKey{capability: ci.GetCapability(), key: key}] = ci
	}
	return m
}

func cover(pending map[string]bool, ci *cpb.CapabilityInfo) (covered []string) {
	for _, p := range ci.Path {
		var key string
		switch *granularity {
		case "package", "intermediate":
			key = p.GetPackage()
		case "function", "":
			key = p.GetName()
		}
		if key == "" {
			continue
		}
		if pending[key] {
			covered = append(covered, key)
			pending[key] = false
		}
	}
	sort.Strings(covered)
	return covered
}

func sortAndPrintCapabilities(cs []cpb.Capability) {
	slices.Sort(cs)
	tw := tabwriter.NewWriter(
		os.Stdout, // output
		10,        // minwidth
		8,         // tabwidth
		4,         // padding
		' ',       // padchar
		0)         // flags
	capabilityDescription := map[cpb.Capability]string{
		2:  "Access to the file system",
		3:  "Access to the network",
		4:  "Read or modify settings in the Go runtime",
		5:  "Read system information, e.g. environment variables",
		6:  "Modify system information, e.g. environment variables",
		7:  `Call miscellaneous functions in the "os" package `,
		8:  "Make system calls",
		9:  "Invoke arbitrary code, e.g. assembly or go:linkname",
		10: "Call cgo functions",
		11: "Code that Capslock cannot effectively analyze",
		12: "Uses unsafe.Pointer",
		13: "Uses reflect",
		14: "Execute other programs, usually via os/exec",
	}
	for _, c := range cs {
		fmt.Fprint(tw, "\t", cpb.Capability_name[int32(c)], ":\t", capabilityDescription[c], "\n")
	}
	tw.Flush()
}

func summarizeNewCapabilities(keys []mapKey, baselineMap, currentMap capabilitiesMap) (newlyUsedCapabilities, existingCapabilitiesWithNewUses []cpb.Capability) {
	hasAnyOldUse := make(map[cpb.Capability]bool)
	newUses := make(map[cpb.Capability]int)
	for _, key := range keys {
		_, inBaseline := baselineMap[key]
		_, inCurrent := currentMap[key]
		if inBaseline {
			hasAnyOldUse[key.capability] = true
		}
		if !inBaseline && inCurrent {
			newUses[key.capability]++
		}
	}
	newUsesOfExistingCapabilities := 0
	for c, n := range newUses {
		if !hasAnyOldUse[c] {
			newlyUsedCapabilities = append(newlyUsedCapabilities, c)
		} else {
			existingCapabilitiesWithNewUses = append(existingCapabilitiesWithNewUses, c)
			newUsesOfExistingCapabilities += n
		}
	}
	if n := len(newlyUsedCapabilities); n > 0 {
		if n == 1 {
			fmt.Println("\nAdded 1 new capability:")
		} else {
			fmt.Printf("\nAdded %d new capabilities:\n", n)
		}
		sortAndPrintCapabilities(newlyUsedCapabilities)
	}
	if n := newUsesOfExistingCapabilities; n > 0 {
		if n == 1 {
			fmt.Println("\nAdded 1 new use of existing capability:")
		} else {
			fmt.Printf("\nAdded %d new uses of existing capabilities:\n", n)
		}
		sortAndPrintCapabilities(existingCapabilitiesWithNewUses)
	}
	if len(newlyUsedCapabilities) == 0 && newUsesOfExistingCapabilities == 0 {
		switch *granularity {
		case "package":
			fmt.Printf("\nBetween those commits, none of those packages gained a new capability.\n")
		case "intermediate":
			fmt.Printf("\nBetween those commits, there were no uses of capabilities via a new package.\n")
		case "function", "":
			fmt.Printf("\nBetween those commits, no functions in those packages gained a new capability.\n")
		}
	}
	return newlyUsedCapabilities, existingCapabilitiesWithNewUses
}

func diffCapabilityInfoLists(baseline, current *cpb.CapabilityInfoList) (different bool) {
	granularityDescription := map[string]string{
		"package":      "Package",
		"intermediate": "Package",
		"function":     "Function",
		"":             "Function",
	}[*granularity]
	baselineMap := populateMap(baseline, *granularity)
	currentMap := populateMap(current, *granularity)
	var keys []mapKey
	for k := range baselineMap {
		keys = append(keys, k)
	}
	for k := range currentMap {
		if _, ok := baselineMap[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if a, b := keys[i].capability, keys[j].capability; a != b {
			return a < b
		}
		return keys[i].key < keys[j].key
	})
	newlyUsedCapabilities, existingCapabilitiesWithNewUses :=
		summarizeNewCapabilities(keys, baselineMap, currentMap)
	// Output changes for each capability, in the order they were printed above.
	for _, list := range [][]cpb.Capability{newlyUsedCapabilities, existingCapabilitiesWithNewUses} {
		for _, c := range list {
			switch *granularity {
			case "package":
				fmt.Printf("\nNew packages with capability %s:\n", c)
			case "intermediate":
				fmt.Printf("\nNew packages in call paths to capability %s:\n", c)
			case "function":
				fmt.Printf("\nNew functions with capability %s:\n", c)
			}

			pending := make(map[string]bool)
			for _, key := range keys {
				if key.capability != c {
					continue
				}
				_, inBaseline := baselineMap[key]
				_, inCurrent := currentMap[key]
				if !inBaseline && inCurrent {
					pending[key.key] = true
					different = true
				}
			}
			for _, key := range keys {
				if key.capability != c {
					continue
				}
				if !pending[key.key] {
					// already done
					continue
				}
				ci := currentMap[key]
				if keys := cover(pending, ci); len(keys) > 1 {
					// This call path can be the example for multiple keys.
					fmt.Printf("\n%ss %s have capability %s:\n", granularityDescription, strings.Join(keys, ", "), key.capability)
				} else {
					fmt.Printf("\n%s %s has capability %s:\n", granularityDescription, key.key, key.capability)
				}
				printCallPath(ci.Path)
			}
		}
	}
	return different
}

func printCallPath(fns []*cpb.Function) {
	tw := tabwriter.NewWriter(
		os.Stdout, // output
		10,        // minwidth
		8,         // tabwidth
		2,         // padding
		' ',       // padchar
		0)         // flags
	for _, f := range fns {
		if f.Site != nil {
			fmt.Fprint(tw, f.Site.GetFilename(), ":", f.Site.GetLine(), ":", f.Site.GetColumn())
		}
		fmt.Fprint(tw, "\t", f.GetName(), "\n")
	}
	tw.Flush()
}
