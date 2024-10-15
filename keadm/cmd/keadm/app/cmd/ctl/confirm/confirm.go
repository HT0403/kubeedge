/*
Copyright 2024 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package confirm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/keadm/cmd/keadm/app/cmd/common"
	"github.com/kubeedge/kubeedge/keadm/cmd/keadm/app/cmd/ctl/client"
)

var (
	edgeConfirmShortDescription = `Send a confirmation signal to the MetaService API`
)

type NodeConfirmOptions struct {
	UpgradeJobName string
}

// NewEdgeConfirm returns KubeEdge confirm command.
func NewEdgeConfirm() *cobra.Command {
	nodeConfirmOptions := NewNodeConfirmOpts()
	cmd := &cobra.Command{
		Use:   "confirm",
		Short: edgeConfirmShortDescription,
		Long:  edgeConfirmShortDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) <= 0 {
				return fmt.Errorf("no node specified for confirm upgrade")
			}
			cmdutil.CheckErr(nodeConfirmOptions.confirmNode(args[0]))
			return nil
		},
	}
	AddNodeConfirmFlags(cmd, nodeConfirmOptions)
	return cmd
}
func (o *NodeConfirmOptions) confirmNode(nodeHostName string) error {
	kubeClient, err := client.KubeClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	confirmResponse, err := nodeConfirm(ctx, kubeClient, o.UpgradeJobName, nodeHostName)
	if err != nil {
		return err
	}

	for _, logMsg := range confirmResponse.LogMessages {
		fmt.Println(logMsg)
	}

	for _, errMsg := range confirmResponse.ErrMessages {
		fmt.Println(errMsg)
	}
	return nil
}
func NewNodeConfirmOpts() *NodeConfirmOptions {
	nodeConfirmOptions := &NodeConfirmOptions{}
	nodeConfirmOptions.UpgradeJobName = "default"
	return nodeConfirmOptions
}

func AddNodeConfirmFlags(cmd *cobra.Command, nodeConfirmOptions *NodeConfirmOptions) {
	cmd.Flags().StringVarP(&nodeConfirmOptions.UpgradeJobName, common.FlagNameConfirmUpgradeJobName, "n", nodeConfirmOptions.UpgradeJobName,
		"NodeUpgradeJob name for upgrade confirm")
}

func nodeConfirm(ctx context.Context, clientSet *kubernetes.Clientset, upgradejobName string, nodeHostName string) (*types.RestartResponse, error) {
	bodyBytes, err := json.Marshal(nodeHostName)
	if err != nil {
		return nil, err
	}
	result := clientSet.CoreV1().RESTClient().Post().
		Resource("taskupgrade" + "/" + upgradejobName).
		SubResource("confirm-upgrade").
		Body(bodyBytes).
		Do(ctx)

	if result.Error() != nil {
		return nil, result.Error()
	}

	statusCode := -1
	result.StatusCode(&statusCode)
	if statusCode != 200 {
		return nil, fmt.Errorf("node upgrade confirm failed with status code: %d", statusCode)
	}

	body, err := result.Raw()
	if err != nil {
		return nil, err
	}

	var restartResponse types.RestartResponse
	err = json.Unmarshal(body, &restartResponse)
	if err != nil {
		return nil, err
	}
	return &restartResponse, nil
}
