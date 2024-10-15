package upgrade

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"k8s.io/klog/v2"

	"github.com/kubeedge/kubeedge/common/types"
	"github.com/kubeedge/kubeedge/edge/pkg/common/dbm"
)

// NodeTaskReq Table
const NodeTaskReqTableName = "nodetask_req"

// NodeTaskRequestTable the struct of NodeTaskRequwst
type NodeTaskRequestTable struct {
	ID    int64  `orm:"column(id);size(64);auto;pk"`
	Key   string `orm:"column(key); size(256); pk"`
	Name  string `orm:"column(name); size(256)"`
	Value string `orm:"column(value);null;type(text)"`
}

func InitNodeTaskRequestTable() orm.Ormer {
	orm.RegisterModel(new(NodeTaskRequestTable))
	obm := dbm.DefaultOrmFunc()
	return obm
}

// QueryNodeTaskRequest query NodeTaskRequest
func QueryNodeTaskRequest(key string, condition string) (*[]NodeTaskRequestTable, error) {
	twin := new([]NodeTaskRequestTable)
	_, err := dbm.DBAccess.QueryTable(NodeTaskReqTableName).Filter(key, condition).All(twin)
	if err != nil {
		return nil, err
	}
	return twin, nil
}

// SaveDeviceTwin save NodeTaskRequestField
func SaveNodeTaskRequestField(o orm.Ormer, doc *NodeTaskRequestTable) error {
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

// SaveNodeTaskRequest save struct of NodeTaskRequest
func SaveNodeTaskRequest(o orm.Ormer, nodetaskreq types.NodeTaskRequest) error {
	var metadata *NodeTaskRequestTable
	nodeTaskReqType := reflect.TypeOf(nodetaskreq)
	nodeTaskReqValue := reflect.ValueOf(nodetaskreq)
	for i := 0; i < nodeTaskReqType.NumField(); i++ {
		fieldType := nodeTaskReqType.Field(i)
		metadata = &NodeTaskRequestTable{
			Key:   strconv.Itoa(i),
			Name:  fieldType.Name,
			Value: nodeTaskReqValue.FieldByName(fieldType.Name).String(),
		}
		err := SaveNodeTaskRequestField(o, metadata)
		if err != nil {
			return fmt.Errorf("Save NodeTaskRequest to DB error:%v", err)
		}
	}
	return nil
}

// DeleteNodeTaskReq delete NodeTaskRequest
func DeleteNodeTaskRequestField(o orm.Ormer, key string, name string) error {
	err := o.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// Delete data
		// Using txOrm to execute SQL
		_, e := txOrm.QueryTable(NodeTaskReqTableName).Filter("key", key).Filter("name", name).Delete()
		// if e != nil the transaction will be rollback
		// or it will be committed
		return e
	})

	if err != nil {
		klog.Errorf("Something wrong when deleting NodeTaskRequest data: %v", err)
		return err
	}
	klog.V(4).Info("Delete NodeTaskRequest data successfully")
	return nil
}

// DeleteNodeTaskRequest delete struct of NodeTaskRequest
func DeleteNodeTaskRequest(o orm.Ormer, nodetaskreq types.NodeTaskRequest) error {
	nodeTaskReqType := reflect.TypeOf(nodetaskreq)
	for i := 0; i < nodeTaskReqType.NumField(); i++ {
		fieldType := nodeTaskReqType.Field(i)
		err := DeleteNodeTaskRequestField(o, strconv.Itoa(i), fieldType.Name)
		if err != nil {
			return fmt.Errorf("Delete NodeTaskRequest to DB error:%v", err)
		}
	}
	return nil
}
