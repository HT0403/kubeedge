package handlerfactory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"reflect"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/kubeedge/kubeedge/common/types"
	commontypes "github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/edgehub/task/taskexecutor"
	v2 "github.com/kubeedge/kubeedge/edge/pkg/metamanager/dao/v2"
	"github.com/kubeedge/kubeedge/edge/pkg/metamanager/metaserver/common"
	"github.com/kubeedge/kubeedge/pkg/version"
	"k8s.io/klog/v2"
)

func (f *Factory) Restart(namespace string) http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		podNameBytes, err := limitedReadBody(req, int64(3*1024*1024))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var podNames []string
		err = json.Unmarshal(podNameBytes, &podNames)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		restartInfo := common.RestartInfo{
			PodNames:  podNames,
			Namespace: namespace,
		}
		restartResponse := f.storage.Restart(req.Context(), restartInfo)
		restartResBytes, err := json.Marshal(restartResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(restartResBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	return h
}
func SaveTaskReq(o orm.Ormer, doc *v2.MetaV2) error {
	err := o.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// insert data
		// Using txOrm to execute SQL
		_, e := txOrm.Insert(doc)
		// if e != nil the transaction will be rollback
		// or it will be committed
		return e
	})
	if err != nil {
		klog.Errorf("Something wrong when insert NodeTaskRequest data: %v", err)
		return err
	}
	klog.V(4).Info("insert NodeTaskRequest data successfully")
	return nil
}
func (f *Factory) ConfirmUpgrade(upgradeJobName string) http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		klog.Infof("Begin to run upgrade command")
		var upgradeReq commontypes.NodeUpgradeJobRequest
		var configFile string
		var nodeTaskReq types.NodeTaskRequest
		var metadata *v2.MetaV2
		ormdb := orm.NewOrm()
		nodeTaskReqType := reflect.TypeOf(nodeTaskReq)
		nodeTaskReqValue := reflect.ValueOf(nodeTaskReq)
		for i := 0; i < nodeTaskReqType.NumField(); i++ {
			fieldType := nodeTaskReqType.Field(i)
			metadata = &v2.MetaV2{
				Key:   strconv.Itoa(i),
				Name:  fieldType.Name,
				Value: nodeTaskReqValue.FieldByName(fieldType.Name).String(),
			}
			err := SaveTaskReq(ormdb, metadata)
			if err != nil {
				klog.Error("Save NodeTaskRequest to DB error!")
				return
			}
		}
		upgradeCmd := fmt.Sprintf("keadm upgrade edge --upgradeID %s --historyID %s --fromVersion %s --toVersion %s --config %s --image %s > /tmp/keadm.log 2>&1",
			upgradeReq.UpgradeID, upgradeReq.HistoryID, version.Get(), upgradeReq.Version, configFile, upgradeReq.Image)

		executor, _ := taskexecutor.GetExecutor(taskexecutor.TaskUpgrade)
		event, _ := executor.Do(nodeTaskReq)
		klog.Info("Confirm Upgrade:" + event.Type + "," + event.Msg)
		// run upgrade cmd to upgrade edge node
		// use nohup command to start a child progress
		command := fmt.Sprintf("nohup %s &", upgradeCmd)
		cmd := exec.Command("bash", "-c", command)
		s, err := cmd.CombinedOutput()
		if err != nil {
			http.Error(w, fmt.Sprintf("run upgrade command %s failed: %v, res: %s", command, err, s),
				http.StatusInternalServerError)
			return
		}
		klog.Infof("!!! Finish upgrade from Version %s to %s ...", version.Get(), upgradeReq.Version)
	})
	return h
}
