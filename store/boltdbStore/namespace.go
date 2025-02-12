/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package boltdbStore

import (
	"errors"
	"fmt"
	"github.com/polarismesh/polaris-server/common/model"
	"sort"
	"strings"
	"time"
)

const tblNameNamespace = "namespace"

type namespaceStore struct {
	handler BoltHandler
}

// AddNamespace 保存一个命名空间
func (n *namespaceStore) AddNamespace(namespace *model.Namespace) error {
	if namespace.Name == "" || namespace.Owner == "" || namespace.Token == "" {
		return errors.New("store add namespace some param are empty")
	}
	return n.handler.SaveValue(tblNameNamespace, namespace.Name, namespace)
}

// UpdateNamespace 更新命名空间
func (n *namespaceStore) UpdateNamespace(namespace *model.Namespace) error {
	if namespace.Name == "" || namespace.Owner == "" {
		return errors.New("store update namespace some param are empty")
	}
	properties := make(map[string]interface{})
	properties["Owner"] = namespace.Owner
	properties["Comment"] = namespace.Comment
	properties["ModifyTime"] = time.Now()
	return n.handler.UpdateValue(tblNameNamespace, namespace.Name, properties)
}

// UpdateNamespaceToken 更新命名空间token
func (n *namespaceStore) UpdateNamespaceToken(name string, token string) error {
	if name == "" || token == "" {
		return fmt.Errorf("Update Namespace Token missing some params")
	}
	properties := make(map[string]interface{})
	properties["Token"] = token
	properties["ModifyTime"] = time.Now()
	return n.handler.UpdateValue(tblNameNamespace, name, properties)
}

// ListNamespaces 查询owner下所有的命名空间
func (n *namespaceStore) ListNamespaces(owner string) ([]*model.Namespace, error) {
	if owner == "" {
		return nil, errors.New("store lst namespaces owner is empty")
	}
	values, err := n.handler.LoadValuesByFilter(
		tblNameNamespace, []string{"Owner"}, &model.Namespace{}, func(value map[string]interface{}) bool {
			ownerValue, ok := value["Owner"]
			if !ok {
				return false
			}
			return strings.Contains(ownerValue.(string), owner)
		})
	if nil != err {
		return nil, err
	}
	return toNamespaces(values), nil
}

// GetNamespace 根据name获取命名空间的详情
func (n *namespaceStore) GetNamespace(name string) (*model.Namespace, error) {
	values, err := n.handler.LoadValues(tblNameNamespace, []string{name}, &model.Namespace{})
	if nil != err {
		return nil, err
	}
	nsValue, ok := values[name]
	if !ok {
		return nil, nil
	}
	ns := nsValue.(*model.Namespace)
	if !ns.Valid {
		return nil, nil
	}
	return ns, nil
}

type NamespaceSlice []*model.Namespace

// Len 命名空间列表长度
func (ns NamespaceSlice) Len() int {
	return len(ns)
}

// Less 比较大小
func (ns NamespaceSlice) Less(i, j int) bool {
	return ns[i].ModifyTime.Before(ns[j].ModifyTime)
}

// Swap 交换元素的位置
func (ns NamespaceSlice) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

// GetNamespaces 从数据库查询命名空间
func (n *namespaceStore) GetNamespaces(
	filter map[string][]string, offset, limit int) ([]*model.Namespace, uint32, error) {
	values, err := n.handler.LoadValuesAll(tblNameNamespace, &model.Namespace{})
	if nil != err {
		return nil, 0, err
	}
	namespaces := NamespaceSlice(toNamespaces(values))
	sort.Sort(sort.Reverse(namespaces))
	startIdx := offset * limit
	if startIdx >= len(namespaces) {
		return nil, 0, nil
	}
	endIdx := startIdx + limit
	if endIdx > len(namespaces) {
		endIdx = len(namespaces)
	}
	return namespaces[startIdx:endIdx], 0, nil
}

func toNamespaces(values map[string]interface{}) []*model.Namespace {
	namespaces := make([]*model.Namespace, 0, len(values))
	for _, nsValue := range values {
		namespaces = append(namespaces, nsValue.(*model.Namespace))
	}
	return namespaces
}

// GetMoreNamespaces 获取增量数据
func (n *namespaceStore) GetMoreNamespaces(mtime time.Time) ([]*model.Namespace, error) {
	values, err := n.handler.LoadValuesByFilter(
		tblNameNamespace, []string{"ModifyTime"}, &model.Namespace{}, func(value map[string]interface{}) bool {
			mTimeValue, ok := value["ModifyTime"]
			if !ok {
				return false
			}
			return mTimeValue.(time.Time).After(mtime)
		})
	if nil != err {
		return nil, err
	}
	return toNamespaces(values), nil
}
