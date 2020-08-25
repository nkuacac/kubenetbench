package core

import (
	"fmt"
	"log"
	"time"

	"github.com/kkourt/kubenetbench/utils"
)

// KubeGetPodIP returns the IP address of a pod using a proper selector
func (c *RunBenchCtx) KubeGetPodIP(
	selector string,
	retries uint,
	st time.Duration,
) (string, error) {

	retriesOrig := retries
	cmd := fmt.Sprintf(
		"kubectl get pod -l \"%s\" -o custom-columns=IP:.status.podIP --no-headers",
		selector,
	)
	for {
		log.Printf("$ %s # (remaining retries: %d)", cmd, retries)
		lines, err := utils.ExecCmdLines(cmd)
		if err == nil && len(lines) == 1 && lines[0] != "<none>" {
			return lines[0], nil
		}

		if retries == 0 {
			return "", fmt.Errorf("Error executing %s after %d retries (last error:%w)", cmd, retriesOrig, err)
		}

		retries--
		time.Sleep(st)
	}
}

// KubeGetPodPhase returns the phase of a pod
func (c *RunBenchCtx) KubeGetPodPhase(selector string) (string, error) {
	cmd := fmt.Sprintf(
		"kubectl get pod -l \"%s\" -o custom-columns=Status:.status.phase --no-headers",
		selector,
	)

	lines, err := utils.ExecCmdLines(cmd)
	if err != nil {
		return "", fmt.Errorf("command %s failed: %w", cmd, err)
	}

	if len(lines) != 1 {
		log.Fatal("Selector did not provide single result cmd:", cmd, "result:", lines)
		return "", fmt.Errorf("selector %s did not provide single result: command: %s; result: %s", selector, cmd, lines)
	}

	return lines[0], nil
}

// KubeGetPodName returns the name of a pod
func (c *RunBenchCtx) KubeGetPodName(selector string) (string, error) {
	cmd := fmt.Sprintf(
		`kubectl get pod -l "%s"  -o custom-columns=Name:.metadata.name --no-headers`,
		selector,
	)

	lines, err := utils.ExecCmdLines(cmd)
	if err != nil {
		return "", fmt.Errorf("command %s failed: %w", cmd, err)
	}

	if len(lines) != 1 {
		log.Fatal("Selector did not provide single result cmd:", cmd, "result:", lines)
		return "", fmt.Errorf("selector %s did not provide single result: command: %s; result: %s", selector, cmd, lines)
	}

	return lines[0], nil
}

// KubeSaveLogs saves logs using a selector
// NB: for whaterver reason, kubecutl logs -l ??? truncates the logs
func (c *RunBenchCtx) KubeSaveLogs(selector string, logfile string) error {
	podname, err := c.KubeGetPodName(selector)
	if err != nil {
		return fmt.Errorf("Failed to get pod name: %w", err)
	}
	argcmd := fmt.Sprintf(`kubectl logs %s > %s`, podname, logfile)
	log.Printf("$ %s ", argcmd)
	return utils.ExecCmd(argcmd)
}

// KubeGetServiceIP returns the ip of a service
// NB: probably a better option to use DNS
func (c *RunBenchCtx) KubeGetServiceIP(
	selector string,
	retries uint,
	st time.Duration,
) (string, error) {

	retriesOrig := retries
	cmd := fmt.Sprintf(
		"kubectl get service -l '%s' -o custom-columns=IP:.spec.clusterIP --no-headers",
		selector,
	)

	for {
		log.Printf("$ %s # (remaining retries: %d)", cmd, retries)
		lines, err := utils.ExecCmdLines(cmd)
		if err == nil && len(lines) == 1 && lines[0] != "<none>" {
			return lines[0], nil
		}

		if retries == 0 {
			return "", fmt.Errorf("Error executing %s after %d retries (last error:%w)", cmd, retriesOrig, err)
		}

		retries--
		time.Sleep(st)
	}
}

// KubeApply calls kubectl apply -f
func (c *RunBenchCtx) KubeApply(fname string) error {
	cmd := fmt.Sprintf("kubectl apply -f %s", fname)
	log.Printf("$ %s ", cmd)
	return utils.ExecCmd(cmd)
}

// KubeCleanup deletes pods and networkpolicies from our run
func (c *RunBenchCtx) KubeCleanup() error {
	cmd := fmt.Sprintf("kubectl delete pod,deployment,service,networkpolicy -l \"knb-runid=%s\"", c.runid)
	log.Printf("$ %s ", cmd)

	if c.cleanup {
		return utils.ExecCmd(cmd)
	} else {
		log.Printf("Cleanup disabled")
	}

	return nil
}