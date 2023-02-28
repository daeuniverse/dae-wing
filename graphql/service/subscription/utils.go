/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package subscription

//
//func ResolveSubscription(log *logrus.Logger, configDir string, subscription string) (tag string, nodes []string, err error) {
//	/// Get tag.
//	tag, subscription = common.GetTagFromLinkLikePlaintext(subscription)
//
//	/// Parse url.
//	u, err := url.Parse(subscription)
//	if err != nil {
//		return tag, nil, fmt.Errorf("failed to parse subscription \"%v\": %w", subscription, err)
//	}
//	log.Debugf("ResolveSubscription: %v", subscription)
//	var (
//		b    []byte
//		resp *http.Response
//	)
//	switch u.Scheme {
//	case "file":
//		b, err = ResolveFile(u, configDir)
//		if err != nil {
//			return "", nil, err
//		}
//		goto resolve
//	default:
//	}
//	resp, err = http.Get(subscription)
//	if err != nil {
//		return "", nil, err
//	}
//	defer resp.Body.Close()
//	b, err = io.ReadAll(resp.Body)
//	if err != nil {
//		return "", nil, err
//	}
//resolve:
//	if nodes, err = ResolveSubscriptionAsSIP008(log, b); err == nil {
//		return tag, nodes, nil
//	} else {
//		log.Debugln(err)
//	}
//	return tag, ResolveSubscriptionAsBase64(log, b), nil
//}
//
//func ImportSubscription(ctx context.Context, link string) (err error) {
//	subscription.ResolveSubscription(logrus.StandardLogger())
//	for _, arg := range argument.Args {
//		m, err := model.NewNodeModel(arg.Link, arg.Remarks, sql.NullInt64{})
//		if err != nil {
//			if errors.Is(err, model.BadLinkFormatError) || argument.IgnoreErrorNode {
//				// Skip this node, but print to log.
//				logrus.WithFields(logrus.Fields{
//					"link": arg.Link,
//					"err":  err,
//				}).Warnf("Failed to import node")
//				continue
//			}
//			// Write error to status instead of returning.
//			m.Status = err.Error()
//		}
//		if err = model.Node.Create(ctx, m); err != nil {
//			return err
//		}
//	}
//	return nil
//}
