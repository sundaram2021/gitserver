package main

import (
	"flag"
	"fmt"
	"gitserver/internal/core"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write CPU profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to file")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	if len(args) < 1 {
		fmt.Println("Usage: mygitserver [command]")
		return
	}
	command := args[0]

	switch command {
	case "init":
		core.InitializeRepository()

	case "add":
		if len(args) < 2 {
			fmt.Println("Usage: mygitserver add [files...]")
			return
		}
		core.AddFile(args[1:])

	case "commit":
		if len(args) < 3 || args[1] != "-m" {
			fmt.Println("Usage: mygitserver commit -m \"message\"")
			return
		}
		core.CommitChanges(args[2:])

	case "branch":
		if len(args) == 1 {
			core.ListBranches()
		} else if len(args) == 2 {
			core.CreateBranch(args[1])
		} else {
			fmt.Println("Usage: mygitserver branch [branch-name]")
		}

	case "checkout":
		if len(args) < 2 {
			fmt.Println("Usage: mygitserver checkout [branch-name]")
			return
		}
		core.SwitchBranch(args[1])

	case "merge":
		if len(args) < 2 {
			fmt.Println("Usage: mygitserver merge [branch]")
			return
		}
		core.MergeBranch(args[1])

	case "rebase":
		fmt.Println("Rebase functionality not implemented yet")

	case "status":
		core.Status()

	case "log":
		core.Log()

	case "diff":
		core.Diff()

	default:
		fmt.Println("Unknown command")
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create memory profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		runtime.GC() 
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not write memory profile: %v\n", err)
			os.Exit(1)
		}
	}
}
