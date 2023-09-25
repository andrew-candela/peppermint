package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func runCommand() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:1053", nil)
	if err != nil {
		return fmt.Errorf("problem constructing publish request... %w", err)
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("problem performing publish request... %w", err)
	}
	return nil
}

func RunTestCommand() {
	err := runCommand()
	if err != nil {
		fmt.Println(err)
	}
}
