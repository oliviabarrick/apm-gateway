package exporter

import (
	apm "go.elastic.co/apm/model"
)

type Exporter interface {
	SendToAPM(*apm.Transaction) error
}
