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
	verbose = flag.Bool("v", false, "enable verbose logging")
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

func populateMap(cil *cpb.CapabilityInfoList) capabilitiesMap {
	m := make(capabilitiesMap)
	for _, ci := range cil.GetCapabilityInfo() {
		key := ci.GetPackageDir()
		if key == "" {
			continue
		}
		m[mapKey{capability: ci.GetCapability(), key: key}] = ci
	}
	return m
}

func cover(pending map[string]bool, ci *cpb.CapabilityInfo) (covered []string) {
	for _, p := range ci.Path {
		key := p.GetPackage()
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
		fmt.Fprint(tw, cpb.Capability_name[int32(c)], ":\t", capabilityDescription[c], "\n")
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
	scope := os.Getenv("SCOPE")
	if scope != "" {
		scope = fmt.Sprintf(" for %s", scope)
	}
	fmt.Printf("## Capability Report%s\n", scope)
	if n := len(newlyUsedCapabilities); n > 0 {
		fmt.Printf("\n### ⚙️ Added %d new Capabilities", n)
		fmt.Printf("\n```\n")
		sortAndPrintCapabilities(newlyUsedCapabilities)
		fmt.Printf("```\n")
	}
	if n := newUsesOfExistingCapabilities; n > 0 {
		fmt.Printf("\n### ⚙️ Added %d new uses of existing capabilities", n)
		fmt.Printf("\n```\n")
		sortAndPrintCapabilities(existingCapabilitiesWithNewUses)
		fmt.Printf("```\n")
	}
	if len(newlyUsedCapabilities) == 0 && newUsesOfExistingCapabilities == 0 {
		fmt.Printf("\nBetween those commits, there were no uses of capabilities via a new package.\n")
	}
	return newlyUsedCapabilities, existingCapabilitiesWithNewUses
}

func diffCapabilityInfoLists(baseline, current *cpb.CapabilityInfoList) (different bool) {
	baselineMap := populateMap(baseline)
	currentMap := populateMap(current)
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
			fmt.Printf("\n### ⚠️ New Packages for %s\n", c)

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
				keys := cover(pending, ci)
				fmt.Printf("<details><summary>📦 %s</summary>\n\n", strings.Join(keys, ", "))
				fmt.Printf("")
				printCallPath(ci.Path)
				fmt.Printf("</details>\n\n")
			}
		}
	}
	return different
}

func printCallPath(fns []*cpb.Function) {
	fmt.Printf("```\n")
	tw := tabwriter.NewWriter(
		os.Stdout, // output
		10,        // minwidth
		8,         // tabwidth
		2,         // padding
		' ',       // padchar
		0)         // flags
	fmt.Printf("- Packages tree:\n")
	for _, f := range fns {
		if f.Site.GetFilename() == "" && !strings.Contains(f.GetName(), "(") {
			fmt.Fprint(tw, "\t", f.GetName(), "\n")
		}
	}
	tw.Flush()
	fmt.Printf("- Path:\n")
	for _, f := range fns {
		if f.Site.GetFilename() != "" || strings.Contains(f.GetName(), "(") {
			if f.Site.GetFilename() != "" {
				fmt.Fprint(tw, "\t", f.Site.GetFilename(), ":", f.Site.GetLine(), ":", f.Site.GetColumn())
				fmt.Fprint(tw, "\t", f.GetName(), "\n")
			} else {
				fmt.Fprint(tw, "\t\t", f.GetName(), "\n")
			}
		}
	}
	tw.Flush()
	fmt.Printf("```\n")
}
