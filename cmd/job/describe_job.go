package job

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai-cli/cmd/completion"
	"github.com/run-ai/runai-cli/pkg/authentication/assertion"
	commandUtil "github.com/run-ai/runai-cli/pkg/util/command"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/run-ai/runai-cli/cmd/constants"
	"github.com/run-ai/runai-cli/cmd/flags"
	"github.com/run-ai/runai-cli/cmd/trainer"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/types"
	"github.com/run-ai/runai-cli/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type eventAndName struct {
	event v1.Event
	name  string
	index int
}

type PrintArgs struct {
	ShowEvents bool
	Output     string
}

func RunDescribeJobDEPRECATED(cmd *cobra.Command, printArgs PrintArgs, name string) {

	kubeClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	clientset := kubeClient.GetClientset()
	namespace, err := flags.GetNamespaceToUseFromProjectFlag(cmd, kubeClient)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	job, err := trainer.SearchTrainingJob(kubeClient, name, "", namespace)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	printTrainingJob(clientset, job, printArgs)
}

func DescribeCommand() *cobra.Command {
	printArgs := PrintArgs{}
	var command = &cobra.Command{
		Use:     "job JOB_NAME",
		Aliases: []string{"jobs"},
		Short:   "Display details of a job.",
		ValidArgsFunction: GenJobNames,
		PreRun:  commandUtil.RoleAssertion(assertion.AssertViewerRole),
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name := args[0]
			job, clientSet, err := PrepareJobInfo(cmd, name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			printTrainingJob(clientSet, job, printArgs)
		},
	}

	command.Flags().BoolVarP(&printArgs.ShowEvents, "events", "e", true, "Show events relating to job lifecycle.")

	command.Flags().StringVarP(&printArgs.Output, "output", "o", "", "Output format. One of: json|yaml|wide")
	command.RegisterFlagCompletionFunc("output", completion.OutputFormatValues)

	command.Flags().MarkDeprecated("events", "default is true")
	return command
}

func PrepareJobInfo(cmd *cobra.Command, name string) (trainer.TrainingJob, kubernetes.Interface, error) {
	kubeClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	namespace, err := flags.GetNamespaceToUseFromProjectFlag(cmd, kubeClient)
	if err != nil {
		return nil, nil, err
	}

	clientSet := kubeClient.GetClientset()

	job, err := trainer.SearchTrainingJob(kubeClient, name, "", namespace)
	return job, clientSet, err
}

func printTrainingJob(client kubernetes.Interface, job trainer.TrainingJob, printArgs PrintArgs) {
	switch printArgs.Output {
	case "name":
		fmt.Println(job.Name())
		// for future CRD support
	case "json":
		outBytes, err := json.MarshalIndent(BuildJobInfo(job, client), "", "    ")
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Println(string(outBytes))
		}
	case "yaml":
		outBytes, err := yaml.Marshal(BuildJobInfo(job, client))
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Println(string(outBytes))
		}
	case "wide", "":
		printSingleJobHelper(client, job, printArgs)
	default:
		log.Fatalf("Unknown output format: %s", printArgs.Output)
	}
}

func printSingleJobHelper(client kubernetes.Interface, job trainer.TrainingJob, printArgs PrintArgs) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	printJobSummary(w, job)

	// apply a dummy FgDefault format to align tabwriter with the rest of the columns
	fmt.Fprintf(w, "Pods:\n")
	fmt.Fprintf(w, "POD\tSTATUS\tTYPE\tAGE\tNODE\n")
	pods := job.AllPods()

	for _, pod := range pods {
		// hostIP := "N/A"

		var hostIP string
		if pod.Spec.NodeName != "" {
			hostIP = pod.Spec.NodeName + "/" + pod.Status.HostIP
		} else {
			hostIP = pod.Status.HostIP
		}
		// if pod.Status.Phase == v1.PodRunning {
		// }

		if len(hostIP) == 0 {
			hostIP = "N/A"
		}

		podStatus, ok := pod.Annotations[constants.WorkloadCalculatedStatus]
		if !ok {
			podStatus = string(pod.Status.Phase)
		}
		var podCreationTime time.Duration
		if pod.CreationTimestamp.IsZero() {
			podCreationTime = 0
		}

		podCreationTime = metav1.Now().Sub(pod.CreationTimestamp.Time)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", pod.Name,
			strings.ToUpper(podStatus),
			strings.ToUpper(job.Trainer()),
			util.ShortHumanDuration(podCreationTime),
			hostIP)
	}

	if printArgs.ShowEvents {
		printEvents(client, w, job.Namespace(), job)
	}

	_ = w.Flush()

}

func printJobSummary(w io.Writer, job trainer.TrainingJob) {
	fmt.Fprintf(w, "NAME: %s\n", job.Name())
	fmt.Fprintf(w, "NAMESPACE: %s\n", job.Namespace())
	fmt.Fprintf(w, "TYPE: %s\n", job.Trainer())
	fmt.Fprintf(w, "STATUS: %s\n", GetJobRealStatus(job))
	fmt.Fprintf(w, "TRAINING DURATION: %s\n", util.ShortHumanDuration(job.Duration()))
	fmt.Fprintf(w, "GPUS: %v\n", job.RequestedGPUString())
	fmt.Fprintf(w, "TOTAL REQUESTED GPUS: %v\n", job.TotalRequestedGPUsString())
	fmt.Fprintf(w, "ALLOCATED GPUS: %g\n", job.CurrentAllocatedGPUs())
	fmt.Fprintf(w, "ALLOCATED GPUS MEMORY: %v\n", job.CurrentAllocatedGPUsMemory())
	fmt.Fprintf(w, "RUNNING PODS: %d\n", job.RunningPods())
	fmt.Fprintf(w, "PENDING PODS: %d\n", job.PendingPods())
	fmt.Fprintf(w, "PARALLELISM: %d\n", job.Parallelism())
	fmt.Fprintf(w, "COMPLETIONS: %d\n", job.Completions())
	fmt.Fprintf(w, "SUCCEEDED PODS: %d\n", job.Succeeded())
	fmt.Fprintf(w, "FAILED PODS: %d\n", job.Failed())
	fmt.Fprintf(w, "IS DISTRIBUTED WORKLOAD: %s\n", strconv.FormatBool(job.WorkloadType() == "MPIJob"))
	fmt.Fprintf(w, "CREATED BY CLI: %s\n", strconv.FormatBool(job.CreatedByCLI()))
	fmt.Fprintf(w, "SERVICE URL(S): %s\n", strings.Join(job.ServiceURLs(), ", "))
	fmt.Fprintln(w, "")

}

func printEvents(clientset kubernetes.Interface, w io.Writer, namespace string, job trainer.TrainingJob) {
	fmt.Fprintf(w, "\nEvents: \n")
	eventsMap, err := getResourcesEvents(clientset, namespace, job)
	if err != nil {
		fmt.Fprintf(w, "Get job events failed, due to: %v", err)
		return
	}
	if len(eventsMap) == 0 {
		fmt.Fprintln(w, "No events for resources")
		return
	}
	fmt.Fprintf(w, "SOURCE\tTYPE\tAGE\tMESSAGE\n")
	fmt.Fprintf(w, "--------\t----\t---\t-------\n")

	for _, eventAndName := range eventsMap {
		instanceName := fmt.Sprintf("%s/%s", strings.ToLower(eventAndName.event.InvolvedObject.Kind), eventAndName.name)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n",
			instanceName,
			eventAndName.event.Type,
			util.ShortHumanDuration(time.Now().Sub(eventAndName.event.CreationTimestamp.Time)),
			fmt.Sprintf("[%s] %s", eventAndName.event.Reason, eventAndName.event.Message))
		// empty line for per pod
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "", "", "", "", "", "")
	}
}

// Get real job status
// WHen has pods being pending, tfJob still show in Running state, it should be Pending
func GetJobRealStatus(job trainer.TrainingJob) string {
	hasPendingPod := false
	jobStatus := job.GetStatus()
	if jobStatus == "RUNNING" {
		pods := job.AllPods()
		for _, pod := range pods {
			if pod.Status.Phase == v1.PodPending {
				log.Debugf("pod %s is pending", pod.Name)
				hasPendingPod = true
				break
			}
		}
		if hasPendingPod {
			jobStatus = "PENDING"
		}
	}
	return jobStatus
}

// Get Event of the Job
func getResourcesEvents(client kubernetes.Interface, namespace string, job trainer.TrainingJob) ([]eventAndName, error) {
	events, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []eventAndName{}, err
	}

	return getSortedEvents(events.Items, job.Resources(), job.GetPodGroupName()), nil
}

func getSortedEvents(items []v1.Event, resources []types.Resource, podGroupName string) []eventAndName {
	eventAndNames := []eventAndName{}
	index := 0
	for _, event := range items {
		for _, resource := range resources {
			if event.InvolvedObject.Kind == string(resource.ResourceType) && string(event.InvolvedObject.UID) == resource.Uid {
				eventAndNames = append(eventAndNames, eventAndName{event, resource.Name, index})
				index++
				break
			}
		}

		// TODO: We should add pogGroup as a resource of a job and remove this part.
		if len(podGroupName) > 0 && event.InvolvedObject.Name == podGroupName {
			eventAndNames = append(eventAndNames, eventAndName{event, podGroupName, index})
			index++
		}

	}

	sort.Slice(eventAndNames, func(i, j int) bool {
		lv := eventAndNames[i]
		rv := eventAndNames[j]
		if lv.event.CreationTimestamp.Time.Before(rv.event.CreationTimestamp.Time) {
			return true
		}

		if lv.event.CreationTimestamp.Time.After(rv.event.CreationTimestamp.Time) {
			return false

		}

		return lv.index < rv.index
	})

	return eventAndNames
}

/**
* BuildTrainingJobInfo returns types.TrainingJobInfo
 */
func BuildJobInfo(job trainer.TrainingJob, clientset kubernetes.Interface) *types.JobInfo {
	instances := []types.Instance{}
	for _, pod := range job.AllPods() {
		isChief := false
		if pod.Name == job.ChiefPod().Name {
			isChief = true
		}

		instances = append(instances, types.Instance{
			Name:    pod.Name,
			Status:  strings.ToUpper(string(pod.Status.Phase)),
			Age:     util.ShortHumanDuration(job.Age()),
			Node:    pod.Status.HostIP,
			IsChief: isChief,
		})
	}

	return &types.JobInfo{
		Name:      job.Name(),
		Namespace: job.Namespace(),
		Status:    types.JobStatus(GetJobRealStatus(job)),
		Duration:  util.ShortHumanDuration(job.Duration()),
		Trainer:   job.Trainer(),
		Priority:  getPriorityClass(job),
		ChiefName: job.ChiefPod().Name,
		Instances: instances,
	}
}

/**
* getPriorityClass returns priority class name
 */
func getPriorityClass(job trainer.TrainingJob) string {
	pc := job.GetPriorityClass()
	if len(pc) == 0 {
		pc = "N/A"
	}

	return pc
}
