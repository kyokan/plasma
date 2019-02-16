package cmd

import (
	"github.com/spf13/cobra"
	"os/exec"
	"strconv"
	"bufio"
	"fmt"
	"io"
	"strings"
	"os"
	"path/filepath"
	"os/signal"
	"net"
	"time"
	"path"
	"github.com/pkg/errors"
)

const (
	FlagGanacheDBPath    = "ganache-db-path"
	FlagGanachePort      = "ganache-port"
	FlagGanacheBlockTime = "ganache-block-time"
	FlagPlasmaRepoDir    = "plasma-repo-dir"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the harness.",
	RunE:  startHarness,
}

var dbPath string
var ganachePort int
var ganacheBlockTime int
var plasmaRepoDir string

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVar(&dbPath, FlagGanacheDBPath, "/tmp/.ganache", "path to where ganache should store its chain data")
	startCmd.Flags().IntVar(&ganachePort, FlagGanachePort, 8545, "ganache port")
	startCmd.Flags().IntVar(&ganacheBlockTime, FlagGanacheBlockTime, 1, "ganache block time")
	startCmd.Flags().StringVar(&plasmaRepoDir, FlagPlasmaRepoDir, ".", "path to the plasma repository")
}

func startHarness(_ *cobra.Command, _ []string) error {
	var shouldMigrate bool
	p, err := filepath.Abs(dbPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(p); os.IsNotExist(err) {
		err = os.Mkdir(p, 0777)
		if err != nil {
			return err
		}
		shouldMigrate = true
	}

	ganache := exec.Command(
		"ganache-cli",
		"-m",
		"candy maple cake sugar pudding cream honey rich smooth crumble sweet treat",
		"-p",
		strconv.Itoa(ganachePort),
		"--deterministic",
		"--networkId",
		"development",
		"--db",
		dbPath,
		"-b",
		strconv.Itoa(ganacheBlockTime),
	)
	gStdOut, err := ganache.StdoutPipe()
	if err != nil {
		return err
	}
	gStdErr, err := ganache.StderrPipe()
	if err != nil {
		return err
	}

	printPipe("ganache-out", gStdOut)
	printPipe("ganache-err", gStdErr)
	if err := ganache.Start(); err != nil {
		return err
	}

	var started bool
	for i := 0; i < 10; i++ {
		conn, _ := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(ganachePort)), time.Second)
		if conn != nil {
			conn.Close()
			started = true
			break
		}
		time.Sleep(time.Second)
	}

	if !started {
		fmt.Println("ganache didn't start. check the logs and try again. exiting.")
		os.Exit(1)
	}

	if shouldMigrate {
		fmt.Println("Ganache started, migrating...")
		truffle := exec.Command("truffle", "migrate", "--reset")
		truffle.Dir = path.Join(plasmaRepoDir, "plasma-mvp-rootchain")
		tStdOut, err := truffle.StdoutPipe()
		if err != nil {
			return err
		}
		tStdErr, err := truffle.StderrPipe()
		if err != nil {
			return err
		}
		printPipe("truffle-out", tStdOut)
		printPipe("truffle-err", tStdErr)
		if err := truffle.Start(); err != nil {
			return err
		}
		if err := truffle.Wait(); err != nil {
			return err
		}
	} else {
		fmt.Println("Ganache already migrated.")
	}

	fmt.Println("Ganache ready for use.")
	fmt.Println("Contract address: 0xF12b5dd4EAD5F743C6BaA640B0216200e89B60Da")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	return nil
}

func StartGanache(port int, blockTime int, accountCount int, dbPath string) (*exec.Cmd, error) {
	ganache := exec.Command(
		"ganache-cli",
		"-m",
		"candy maple cake sugar pudding cream honey rich smooth crumble sweet treat",
		"-p",
		strconv.Itoa(port),
		"-a",
		strconv.Itoa(accountCount),
		"--deterministic",
		"--networkId",
		"development",
		"--db",
		dbPath,
		"-b",
		strconv.Itoa(blockTime),
	)
	if err := LogCmd(ganache, "ganache"); err != nil {
	    return nil, err
	}
	if err := ganache.Start(); err != nil {
		return nil, err
	}

	var started bool
	for i := 0; i < 10; i++ {
		conn, _ := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
		if conn != nil {
			conn.Close()
			started = true
			break
		}
		time.Sleep(time.Second)
	}

	if !started {
		return nil, errors.New("timeout starting ganache")
	}

	return ganache, nil
}

func MigrateGanache(repoDir string) error {
	truffle := exec.Command("truffle", "migrate", "--reset")
	truffle.Dir = path.Join(repoDir, "plasma-mvp-rootchain")
	if err := LogCmd(truffle, "truffle"); err != nil {
		return err
	}

	if err := truffle.Start(); err != nil {
		return err
	}
	return truffle.Wait()
}

func LogCmd(cmd *exec.Cmd, prefix string) error {
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	printPipe(fmt.Sprintf("%s-out", prefix), stdOut)
	printPipe(fmt.Sprintf("%s-err", prefix), stdErr)
	return nil
}

func printPipe(prefix string, reader io.Reader) {
	go func() {
		str := bufio.NewReader(reader)
		for {
			out, err := str.ReadString('\n')
			if err == io.EOF {
				continue
			}
			if err != nil {
				printPrefix(prefix, err.Error())
				return
			}

			printPrefix(prefix, out)
		}
	}()
}

func printPrefix(prefix string, data string) {
	fmt.Printf("[%s] %s\n", prefix, strings.Trim(data, "\n"))
}
