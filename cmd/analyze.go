package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var analyzeEnv string

type trace struct {
	Time time.Duration

	TraceID      string
	parentSpanID string
	SpanID       string

	Method   string
	Path     string
	FuncName string
}

type traceNode struct {
	Trace    trace
	Children []traceNode
}

type flameNode struct {
	Name     string      `json:"name"`
	Value    int64       `json:"value"`
	Children []flameNode `json:"children"`
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze <project> <funcToAnalyze>",
	Short: "analyze project performance",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}

		funcToAnalyze := args[1]

		logf, err := os.Open(filepath.Join(projectDir, "logs", "access.log"))
		if err != nil {
			log.Fatal(err)
		}
		defer logf.Close()

		span := regexp.MustCompile(`%([\w-]+)/([\w-]+)/([\w-]+)%`)
		timePattern := regexp.MustCompile(`@(\d+\.\d+)`)
		route := regexp.MustCompile(`(?U)\[\d+\] \[.+\] (\w+) (.+) \d+ (.+) `)

		traces := make([]trace, 0)

		scanner := bufio.NewScanner(logf)
		for scanner.Scan() {
			t := trace{}
			line := scanner.Text()
			if !strings.HasPrefix(line, "[200]") {
				continue
			}

			// parse http request and routed function
			r := route.FindAllStringSubmatch(line, -1)
			t.Method = r[0][1]
			t.Path = r[0][2]
			t.FuncName = r[0][3]

			// parse span id
			s := span.FindAllStringSubmatch(line, -1)
			t.TraceID = s[0][1]
			t.parentSpanID = s[0][2]
			if t.parentSpanID == "-" {
				t.parentSpanID = ""
			}
			t.SpanID = s[0][3]

			// parse request time
			second, err := strconv.ParseFloat(timePattern.FindAllStringSubmatch(line, -1)[0][1], 64)
			if err != nil {
				log.Fatal(err)
			}
			ms := int64(second * 1000)
			t.Time = time.Duration(ms) * time.Millisecond

			traces = append(traces, t)
		}
		if err = scanner.Err(); err != nil {
			log.Fatal(err)
		}

		lookup := make(map[string][]trace)
		roots := make([]traceNode, 0)
		for _, t := range traces {
			// if this is not a root span, record it in a map for futher lookup
			if t.parentSpanID != "" {
				if _, ok := lookup[t.parentSpanID]; !ok {
					lookup[t.parentSpanID] = make([]trace, 0)
				}
				lookup[t.parentSpanID] = append(lookup[t.parentSpanID], t)

				continue
			}

			// if this is a root span, check if it's the function we're analyzing
			if t.FuncName != funcToAnalyze {
				continue
			}

			roots = append(roots, traceNode{Trace: t, Children: make([]traceNode, 0)})
		}

		trees := make([]traceNode, 0)
		// get trace tree for each function call
		for _, t := range roots {
			tree := buildTraceTree(lookup, traceNode{Trace: t.Trace, Children: make([]traceNode, 0)})
			trees = append(trees, tree)
		}

		flamegraph := buildFlameTree(trees)

		json, err := json.Marshal(flamegraph[0])
		if err != nil {
			log.Fatal(err)
		}

		ioutil.WriteFile(filepath.Join(projectDir, "temp", "flamegraph.json"), json, os.ModePerm)

		tmplData, err := ioutil.ReadFile(filepath.Join(projectDir, "bin", "flamegraph.html.tmpl"))
		if err != nil {
			log.Fatal(err)
		}
		tmpl, err := template.New("tmp").Parse(string(tmplData))
		if err != nil {
			log.Fatal(err)
		}
		out, err := os.Create(filepath.Join(projectDir, "bin", "flamegraph.html"))
		if err != nil {
			log.Fatal(err)
		}
		err = tmpl.Execute(out, flamegraphTmpl{Data: string(json)})
		if err != nil {
			log.Fatal(err)
		}

		if err := exec.Command("open", "./bin/flamegraph.html").Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

type flamegraphTmpl struct {
	Data string
}

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeEnv, "env", "e", "development", "Environment Name")
	RootCmd.AddCommand(analyzeCmd)
}

func buildTraceTree(lookupTable map[string][]trace, root traceNode) traceNode {
	for _, trace := range lookupTable[root.Trace.SpanID] {
		node := traceNode{Trace: trace, Children: make([]traceNode, 0)}
		root.Children = append(root.Children, buildTraceTree(lookupTable, node))
	}

	return root
}

func buildFlameTree(roots []traceNode) []flameNode {
	flameNodes := make([]flameNode, 0)
	agg := make(map[string][]traceNode)
	// group by function name
	for _, t := range roots {
		if _, ok := agg[t.Trace.FuncName]; !ok {
			agg[t.Trace.FuncName] = make([]traceNode, 0)
		}

		agg[t.Trace.FuncName] = append(agg[t.Trace.FuncName], t)
	}

	// convert to flamenodes

	for funcName, traceNodes := range agg {
		node := flameNode{}
		node.Name = funcName
		next := make([]traceNode, 0)

		for _, t := range traceNodes {
			node.Value += t.Trace.Time.Nanoseconds() / int64(1000000)
			for _, child := range t.Children {
				next = append(next, child)
			}
		}
		node.Value /= int64(len(traceNodes))
		node.Children = buildFlameTree(next)
		flameNodes = append(flameNodes, node)
	}

	return flameNodes
}
