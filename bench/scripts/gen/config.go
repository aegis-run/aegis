package gen

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"os"

	"github.com/spf13/pflag"
)

type Range struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

func (r Range) Roll(rng *rand.Rand) int {
	if r.Max <= r.Min {
		return r.Min
	}
	return r.Min + rng.IntN(r.Max-r.Min+1)
}

type ProbRange struct {
	Range
	IsUserProb float64 `json:"is_user_prob"`
}

type Config struct {
	Seed         []uint64     `json:"seed"`
	Distribution Distribution `json:"distribution"`
	Model        Model        `json:"model"`
	Workload     Workload     `json:"workload"`
	Output       string       `json:"-"`
}

type Distribution struct {
	ZipfS float64 `json:"zipf_s"`
	ZipfV float64 `json:"zipf_v"`
}

type Model struct {
	DeepChainDepth int       `json:"deep_chain_depth"`
	Users          UserSpec  `json:"users"`
	Groups         GroupSpec `json:"groups"`
	Organizations  OrgSpec   `json:"organizations"`
	Directories    DirSpec   `json:"directories"`
	Documents      DocSpec   `json:"documents"`
}

type UserSpec struct {
	Dimension int `json:"dimension"`
}

type GroupSpec struct {
	Dimension int   `json:"dimension"`
	Members   Range `json:"members"`
}

type OrgSpec struct {
	Dimension int   `json:"dimension"`
	Admins    Range `json:"admins"`
	Members   Range `json:"members"`
}

type DirSpec struct {
	Dimension     int       `json:"dimension"`
	Editors       Range     `json:"editors"`
	ParentDirProb float64   `json:"parent_dir_prob"`
	Viewers       ProbRange `json:"viewers"`
}

type DocSpec struct {
	Dimension  int       `json:"dimension"`
	Commenters ProbRange `json:"commenters"`
}

type Workload struct {
	Checks       int     `json:"checks"`
	AllowedRatio float64 `json:"allowed_ratio"`
}

func LoadConfig() Config {
	var configPath string
	var outputPath string
	pflag.StringVar(&configPath, "config", "config.json", "Path to configuration file")
	pflag.StringVar(&outputPath, "output", "", "Path to output directory")
	pflag.Parse()

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("failed to open config file: %v", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("failed to decode config file: %v", err)
	}

	if cfg.Workload.Checks <= 0 {
		log.Fatalf("checks must be greater than 0")
	}

	if cfg.Workload.AllowedRatio < 0 || cfg.Workload.AllowedRatio > 1 {
		log.Fatalf("allowed_ratio must be between 0 and 1")
	}

	if cfg.Distribution.ZipfS <= 1 {
		log.Fatalf("zipf_s must be greater than 1")
	}

	if cfg.Model.Directories.ParentDirProb < 0 || cfg.Model.Directories.ParentDirProb > 1 {
		log.Fatalf("parent_dir_prob must be between 0 and 1")
	}

	cfg.Output = outputPath
	return cfg
}
