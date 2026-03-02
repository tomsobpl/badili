// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package telemetry

import "go.opentelemetry.io/otel/log"

func OtelSeverityFromGelfLevel(level int32) log.Severity {
	switch level {
	case 0: // Emergency
		return log.SeverityFatal4
	case 1: // Alert
		return log.SeverityFatal1
	case 2: // Critical
		return log.SeverityError4
	case 3: // Error
		return log.SeverityError1
	case 4: // Warning
		return log.SeverityWarn
	case 5: // Notice
		return log.SeverityInfo2
	case 6: // Informational
		return log.SeverityInfo1
	case 7: // Debug
		return log.SeverityDebug
	default:
		return log.SeverityUndefined
	}
}
