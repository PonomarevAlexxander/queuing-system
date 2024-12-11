package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"golang.org/x/sync/errgroup"
)

const (
	producer   = "incedent-producer-service"
	processor  = "incedent-processing-service"
	dispatcher = "incedent-dispatcher"

	out = "out"
	src = "src"

	configPath = "config/config.yaml"
)

var (
	producerCmd   = strings.Join([]string{out, producer}, "/")
	dispatcherCmd = strings.Join([]string{out, dispatcher}, "/")
	processorCmd  = strings.Join([]string{out, processor}, "/")

	producerCfg   = strings.Join([]string{src, producer, configPath}, "/")
	dispatcherCfg = strings.Join([]string{src, dispatcher, configPath}, "/")
	processorCfg  = strings.Join([]string{src, processor, configPath}, "/")
)

var args struct {
	Producers  int `arg:"required"`
	Processors int `arg:"required"`
	Pport      int `arg:"required"`
}

func main() {
	arg.MustParse(&args)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dispatcher := createCmd(ctx, dispatcherCmd, "--config", dispatcherCfg)
	if err := dispatcher.Start(); err != nil {
		fmt.Printf("[EXEC] Failed to start dispatcher: %v", err)
	}

	time.Sleep(1 * time.Second)

	var eg errgroup.Group
	eg.Go(dispatcher.Wait)

	for i := 0; i < args.Processors; i++ {
		host := "localhost:" + strconv.Itoa(args.Pport+i)
		eg.Go(
			newCmd().
				with(ctx, processorCmd,
					"--config", processorCfg, "--id", strconv.Itoa(i), "--host", host).
				start,
		)
	}

	for i := 0; i < args.Producers; i++ {
		eg.Go(
			newCmd().
				with(ctx, producerCmd,
					"--config", producerCfg, "--priority", strconv.Itoa(i)).
				start,
		)
	}

	err := eg.Wait()
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	fmt.Println("[EXEC] All services stopped")
}

type cmdWrap struct {
	run func() error
}

func newCmd() *cmdWrap {
	return &cmdWrap{}
}

func (c *cmdWrap) with(ctx context.Context, app string, args ...string) *cmdWrap {
	c.run = func() error {
		return runCmd(ctx, app, args...)
	}

	return c
}

func (c *cmdWrap) start() error {
	return c.run()
}

func runCmd(ctx context.Context, app string, args ...string) error {
	fmt.Printf("[EXEC] Running new cmd [%s] with args [%v]\n", app, args)
	cmd := createCmd(ctx, app, args...)
	return cmd.Run()
}

func createCmd(ctx context.Context, app string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, app, args...)
	if app == dispatcherCmd {
		cmd.Stdout = os.Stdout
	}
	cmd.Cancel = func() error {
		fmt.Printf("[EXEC] Stopping cmd [%s]\n", app)
		return cmd.Process.Signal(syscall.SIGTERM)
	}

	return cmd
}
