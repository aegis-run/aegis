package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aegis-run/aegis/bench/scripts/gen"
)

func main() {
	cfg := gen.LoadConfig()

	if cfg.Output == "" {
		log.Fatal("output directory is required (use --output flag)")
	}

	if err := os.MkdirAll(cfg.Output, 0755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	fmt.Printf("Generating artifacts in %s (seed=%v)...\n", cfg.Output, cfg.Seed)

	// 1. Generate Dataset
	generator := gen.NewGenerator(&cfg)
	artifacts := generator.Generate()
	ds := artifacts.Dataset

	// 2. Build Index & Evaluator
	idx := gen.NewDatasetIndex(ds)
	eval := gen.NewEvaluator(idx)

	// 3. Generate Tuples
	artifacts.Tuples = gen.GenerateTuples(ds)

	// 4. Generate Workload
	artifacts.Checks, artifacts.Herd = gen.GenerateWorkload(generator, ds, idx, eval)

	// 5. Build Metadata
	artifacts.Meta.Config = cfg
	artifacts.Meta.Artifacts.TupleCount = len(artifacts.Tuples)
	artifacts.Meta.Artifacts.CheckCount = len(artifacts.Checks)
	artifacts.Meta.DeepChain = ds.DeepChainFixture

	for _, c := range artifacts.Checks {
		if c.Expected {
			artifacts.Meta.Artifacts.AllowedCount++
		} else {
			artifacts.Meta.Artifacts.DeniedCount++
		}
	}

	// 6. Write Artifacts
	if err := gen.WriteTuplesCSV(filepath.Join(cfg.Output, "tuples.csv"), artifacts.Tuples); err != nil {
		log.Fatalf("failed to write tuples: %v", err)
	}

	if err := gen.WriteChecksJSONL(filepath.Join(cfg.Output, "checks.jsonl"), artifacts.Checks); err != nil {
		log.Fatalf("failed to write checks: %v", err)
	}

	if artifacts.Herd != nil {
		if err := gen.WriteJSON(filepath.Join(cfg.Output, "herd_check.json"), artifacts.Herd); err != nil {
			log.Fatalf("failed to write herd check: %v", err)
		}
	}

	if err := gen.WriteJSON(filepath.Join(cfg.Output, "metadata.json"), artifacts.Meta); err != nil {
		log.Fatalf("failed to write metadata: %v", err)
	}

	fmt.Printf("Done. Generated %d tuples and %d checks.\n", len(artifacts.Tuples), len(artifacts.Checks))
}
