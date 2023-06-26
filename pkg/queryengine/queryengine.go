package queryengine

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"wunderbase/pkg/api"
)

func Run(ctx context.Context, wg *sync.WaitGroup, queryEnginePath, queryEnginePort, prismaSchemaFilePath string, production bool) {
	// when start prisma query engine ,
	// we're not able to listen on the same port,
	// if last engine instance still alive.
	// so we must kill the existing engine process before we start new onw.

	args := []string{"--datamodel-path", prismaSchemaFilePath}
	
	args = append(args, "--host", api.AdditionalConfig.QueryEngineHostBind)

	if !production {
		killExistingPrismaQueryEngineProcess(queryEnginePort)
		args = append(args, "--enable-playground", "--port", queryEnginePort)
	}

	if api.AdditionalConfig.EnableRawQueries {
		args = append(args, "--enable-raw-queries")
	}

	if api.AdditionalConfig.EnableQueryEngineLog {
		args = append(args, "--log-queries")
	}

	if api.AdditionalConfig.EnableMetrics {
		args = append(args, "--enable-metrics", "--dataproxy-metric-override")
	}

	if api.AdditionalConfig.EnableOpenTelemetry {
		args = append(args, "--enable-open-telemetry")
		if api.AdditionalConfig.OpenTelemetryEndpoint != "" {
			args = append(args, "--open-telemetry-endpoint", api.AdditionalConfig.OpenTelemetryEndpoint)
		}
	}

	if api.AdditionalConfig.EnableTelemetryInResponse {
		args = append(args, "--enable-telemetry-in-response")
	}

	cmd := exec.CommandContext(ctx, queryEnginePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalln("run query engine", err)
	}
	<-ctx.Done()
	err = cmd.Process.Kill()
	if err != nil {
		log.Println("kill query engine", err)
	}
	log.Println("Query Engine stopped")
	wg.Done()
}

// reference:https://github.com/wundergraph/wundergraph
func killExistingPrismaQueryEngineProcess(queryEnginePort string) {
	var err error
	if runtime.GOOS == "windows" {
		command := fmt.Sprintf("(Get-NetTCPConnection -LocalPort %s).OwningProcess -Force", queryEnginePort)
		_, err = execCmd(exec.Command("Stop-Process", "-Id", command))
	} else {

		isExistProcessCommand := fmt.Sprintf("lsof -i tcp:%s", queryEnginePort)
		var data []byte

		data, err = execCmd(exec.Command("sh", "-c", isExistProcessCommand))
		if err == nil && len(data) > 0 {
			log.Printf("Found Prisma Query Engine is running on Port %s, Kill It", queryEnginePort)
			command := fmt.Sprintf("lsof -i tcp:%s | grep LISTEN | awk '{print $2}' | xargs kill -9", queryEnginePort)
			data, err = execCmd(exec.Command("sh", "-c", command))
			if err == nil && len(data) > 0 {
				_, err = execCmd(exec.Command("kill", "-9", strings.TrimSpace(string(data))))
			}
		} else {
			log.Printf("Prisma Query Engine is not running on Port %s", queryEnginePort)
			return
		}

	}
	if err != nil {
		var waitStatus syscall.WaitStatus
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			log.Printf("Error killing prisma query (exit code: %d) %s\n", waitStatus.ExitStatus(), err)
		}
	}
}

func execCmd(cmd *exec.Cmd) ([]byte, error) {
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	// Connecting Stderr can help debugging when something goes wrong
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}
