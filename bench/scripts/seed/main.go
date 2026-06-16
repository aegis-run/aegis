package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	datav1 "github.com/aegis-run/aegis/proto/aegis/data/v1"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

func main() {
	input := flag.String("input", "", "Path to tuples.csv")
	output := flag.String("output", "", "Directory to save the consistency token (defaults to input dir)")
	addr := flag.String("addr", "localhost:43615", "Aegis gRPC address")
	token := flag.String("token", "aegis_XBHFNLxp9usfmbAuZtrD0Ajiace1HNfb/CpVLy3qcek=", "Aegis PSK token")
	batchSize := flag.Int("batch", 500, "Mutation batch size")
	concurrency := flag.Int("concurrency", 8, "Number of concurrent workers")
	flag.Parse()

	if *input == "" {
		log.Fatal("--input is required")
	}

	if *output == "" {
		*output = filepath.Dir(*input)
	}

	tuples, err := readTuples(*input)
	if err != nil {
		log.Fatalf("failed to read tuples: %v", err)
	}

	fmt.Printf("Seeding %d tuples via gRPC to %s...\n", len(tuples), *addr)

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := datav1.NewDataClient(conn)
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+*token)

	start := time.Now()
	finalToken, err := seed(ctx, client, tuples, *batchSize, *concurrency)
	if err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("Done in %v (avg %.2f tuples/sec)\n", duration, float64(len(tuples))/duration.Seconds())
	fmt.Printf("Final Consistency Token: %s\n", finalToken)

	if err := os.MkdirAll(*output, 0755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	tokenPath := filepath.Join(*output, "token.txt")
	if err := os.WriteFile(tokenPath, []byte(finalToken), 0644); err != nil {
		log.Printf("failed to write token file: %v", err)
	}
}

type tupleRow struct {
	rt, ri, rel, st, si, sp string
}

func readTuples(path string) ([]tupleRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	// Skip header
	if _, err := r.Read(); err != nil {
		return nil, err
	}

	var tuples []tupleRow
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		tuples = append(tuples, tupleRow{rec[0], rec[1], rec[2], rec[3], rec[4], rec[5]})
	}
	return tuples, nil
}

func seed(ctx context.Context, client datav1.DataClient, tuples []tupleRow, batchSize, concurrency int) (string, error) {
	g, ctx := errgroup.WithContext(ctx)
	tupleChan := make(chan []tupleRow)

	var mu sync.Mutex
	var lastToken string

	// Workers
	for range concurrency {
		g.Go(func() error {
			for batch := range tupleChan {
				mutations := make([]*datav1.TupleMutation, len(batch))
				for i, t := range batch {
					mutations[i] = &datav1.TupleMutation{
						Operation: datav1.TupleMutation_OPERATION_WRITE,
						Tuple: &v1.Tuple{
							Resource: &v1.Instance{Type: t.rt, Id: t.ri},
							Relation: t.rel,
							Subject: &v1.Subject{
								Instance:   &v1.Instance{Type: t.st, Id: t.si},
								Permission: t.sp,
							},
						},
					}
				}

				res, err := client.Mutate(ctx, &datav1.MutateRequest{Mutations: mutations})
				if err != nil {
					return err
				}

				mu.Lock()
				lastToken = res.ConsistencyToken.GetToken()
				mu.Unlock()
			}
			return nil
		})
	}

	// Producer
	g.Go(func() error {
		defer close(tupleChan)
		for i := 0; i < len(tuples); i += batchSize {
			end := i + batchSize
			if end > len(tuples) {
				end = len(tuples)
			}
			select {
			case tupleChan <- tuples[i:end]:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return "", err
	}

	return lastToken, nil
}
