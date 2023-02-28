/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package node

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/v2rayA/dae-wing/model"
)

type ImportNodesArgument struct {
	IgnoreErrorNode bool
	Args            []ImportNodeArgument
}

type ImportNodeArgument struct {
	Link    string
	Remarks string
}

func ImportNode(ctx context.Context, argument ImportNodeArgument) (err error) {
	m, err := model.NewNodeModel(argument.Link, argument.Remarks, sql.NullInt64{})
	if err != nil {
		return fmt.Errorf("failed to parse link: %v")
	}
	if err = model.Node.Create(ctx, m); err != nil {
		return err
	}
	return nil
}

func ImportNodes(ctx context.Context, argument ImportNodesArgument) (err error) {
	for _, arg := range argument.Args {
		m, err := model.NewNodeModel(arg.Link, arg.Remarks, sql.NullInt64{})
		if err != nil {
			if errors.Is(err, model.BadLinkFormatError) || argument.IgnoreErrorNode {
				// Skip this node, but print to log.
				logrus.WithFields(logrus.Fields{
					"link": arg.Link,
					"err":  err,
				}).Warnf("Failed to import node")
				continue
			}
			// Write error to status instead of returning.
			m.Status = err.Error()
		}
		if err = model.Node.Create(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
