package constant

import (
	"time"
)

const (
	ServiceName       = "fox-plugin.npool.top"
	GrpcTimeout       = time.Second * 10
	WaitMsgOutTimeout = time.Second * 40
)
