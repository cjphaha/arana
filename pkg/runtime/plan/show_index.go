/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plan

import (
	"context"
	"io"
	"strings"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/arana-db/arana/pkg/mysql"
	"github.com/arana-db/arana/pkg/proto"
	"github.com/arana-db/arana/pkg/runtime/ast"
)

var _ proto.Plan = (*ShowIndexPlan)(nil)

type ShowIndexPlan struct {
	basePlan
	Stmt *ast.ShowIndex
}

func (s *ShowIndexPlan) Type() proto.PlanType {
	return proto.PlanTypeQuery
}

func (s *ShowIndexPlan) ExecIn(ctx context.Context, conn proto.VConn) (proto.Result, error) {
	var (
		sb      strings.Builder
		indexes []int
		res     proto.Result
		err     error
	)

	if err = s.Stmt.Restore(ast.RestoreDefault, &sb, &indexes); err != nil {
		return nil, errors.WithStack(err)
	}

	var (
		query = sb.String()
		args  = s.toArgs(indexes)
	)

	if res, err = conn.Query(ctx, "", query, args...); err != nil {
		return nil, errors.WithStack(err)
	}

	if closer, ok := res.(io.Closer); ok {
		defer func() {
			_ = closer.Close()
		}()
	}

	return &mysql.Result{
		Fields:   res.GetFields(),
		Rows:     res.GetRows(),
		DataChan: make(chan proto.Row, 1),
	}, nil
}